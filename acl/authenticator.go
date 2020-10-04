package acl

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/gobwas/glob"
	"github.com/oragono/go-ident"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists       = errors.New("user exists")
	ErrUserDoesntExist  = errors.New("user does not exist")
	ErrGroupExists      = errors.New("group exists")
	ErrGroupDoesntExist = errors.New("group does not exist")
)

type AuthenticatorOpts struct {
	DB string `goftpd:"db"`
}

type Authenticator interface {
	// create
	AddUser(string, string) (*User, error)
	AddGroup(string) (*Group, error)

	// get
	GetUser(string) (*User, error)
	GetUsers() ([]*User, error)
	GetGroup(string) (*Group, error)
	GetGroups() ([]*Group, error)

	// save
	UpdateUser(string, func(*User) error) error
	UpdateGroup(string, func(*Group) error) error

	// delete
	DeleteUser(user string) error
	DeleteGroup(group string) error

	// utilities
	CheckPassword(string, string) bool
	CheckIP(string, net.Addr, net.Addr) bool
	ChangePassword(string, string) error
}

// Entry describes an Authenticator Entry
type Entry interface {
	Key() []byte
	SetUpdatedAt()
}

// BadgerAuthenticator implements an Authenticator using a badge key/value store
type BadgerAuthenticator struct {
	db         *badger.DB
	bufferPool sync.Pool
}

// NewBadgerAuthenticator takes in options and a badger DB and returns a new BadgerAuthenticator
// which implements the Authenticator interface
func NewBadgerAuthenticator(db *badger.DB) *BadgerAuthenticator {
	return &BadgerAuthenticator{
		db: db,
		bufferPool: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}
}

func (a *BadgerAuthenticator) encodeAndUpdate(tx *badger.Txn, e Entry) error {
	e.SetUpdatedAt()

	enc := msgpack.GetEncoder()
	defer msgpack.PutEncoder(enc)

	b := a.bufferPool.Get().(*bytes.Buffer)
	b.Reset()
	defer a.bufferPool.Put(b)

	enc.Reset(b)

	if err := enc.Encode(e); err != nil {
		return err
	}

	return tx.Set(e.Key(), b.Bytes())
}

func (a *BadgerAuthenticator) getAndDecode(tx *badger.Txn, key []byte, e Entry) error {
	item, err := tx.Get(key)
	if err != nil {
		return err
	}

	return a.decode(item, e)
}

func (a *BadgerAuthenticator) decode(item *badger.Item, e Entry) error {
	return item.Value(func(val []byte) error {
		dec := msgpack.GetDecoder()
		defer msgpack.PutDecoder(dec)

		dec.ResetBytes(val)

		if err := dec.Decode(e); err != nil {
			return err
		}

		return nil
	})
}

// AddUser creates a user setting the password
func (a *BadgerAuthenticator) AddUser(name, pass string) (*User, error) {
	// check if we have a user by that name
	u, err := a.GetUser(name)
	if err == nil {
		return nil, ErrUserExists
	}

	if err != ErrUserDoesntExist {
		return nil, err
	}

	// hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u = &User{}

	u.Name = name
	u.Password = hashed
	u.CreatedAt = time.Now()

	err = a.db.Update(func(tx *badger.Txn) error {
		return a.encodeAndUpdate(tx, u)
	})
	if err != nil {
		return nil, err
	}

	return u, nil
}

// AddGroup creates a Group
func (a *BadgerAuthenticator) AddGroup(name string) (*Group, error) {
	// check if we have a user by that name
	g, err := a.GetGroup(name)
	if err == nil {
		return nil, ErrGroupExists
	}

	if err != ErrGroupDoesntExist {
		return nil, err
	}

	g = &Group{}

	g.Name = name
	g.CreatedAt = time.Now()

	err = a.db.Update(func(tx *badger.Txn) error {
		return a.encodeAndUpdate(tx, g)
	})
	if err != nil {
		return nil, err
	}

	return g, nil
}

// GetUser attempts to retrieve a User from the store using the name
func (a *BadgerAuthenticator) GetUser(name string) (*User, error) {
	u := &User{Name: name}

	err := a.db.View(func(tx *badger.Txn) error {
		return a.getAndDecode(tx, u.Key(), u)
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrUserDoesntExist
		}
		return nil, err
	}

	return u, nil
}

func (a *BadgerAuthenticator) GetUsers() ([]*User, error) {
	var users []*User

	err := a.db.View(func(tx *badger.Txn) error {
		opts := badger.DefaultIteratorOptions

		opts.PrefetchSize = 10
		opts.Prefix = []byte("users:")
		it := tx.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			var u User
			if err := a.decode(it.Item(), &u); err != nil {
				return err
			}
			users = append(users, &u)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return users, nil
}

// GetGroup attempts to retrieve a Group from the store using the name
func (a *BadgerAuthenticator) GetGroup(name string) (*Group, error) {
	g := &Group{Name: name}

	err := a.db.View(func(tx *badger.Txn) error {
		return a.getAndDecode(tx, g.Key(), g)
	})
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrGroupDoesntExist
		}
		return nil, err
	}

	return g, nil
}

func (a *BadgerAuthenticator) GetGroups() ([]*Group, error) {
	var groups []*Group

	err := a.db.View(func(tx *badger.Txn) error {
		opts := badger.DefaultIteratorOptions

		opts.PrefetchSize = 10
		opts.Prefix = []byte("groups:")
		it := tx.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			var u Group
			if err := a.decode(it.Item(), &u); err != nil {
				return err
			}
			groups = append(groups, &u)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return groups, nil
}

func (a *BadgerAuthenticator) updateEntry(e Entry, fn func(Entry) error) error {
	var count int

	for {
		err := a.db.Update(func(tx *badger.Txn) error {

			if err := a.getAndDecode(tx, e.Key(), e); err != nil {
				if err == badger.ErrKeyNotFound {
					return ErrUserDoesntExist
				}
				return err
			}

			if err := fn(e); err != nil {
				return err
			}

			return a.encodeAndUpdate(tx, e)
		})

		switch err {
		case nil:
			return nil

		case badger.ErrConflict:
			if count > 10 {
				return err
			}
			count++

		default:
			return err
		}

	}

	return nil
}

// UpdateUser overwrites the User in the store
func (a *BadgerAuthenticator) UpdateUser(name string, fn func(*User) error) error {
	u := User{Name: name}
	return a.updateEntry(&u, func(e Entry) error {
		user, ok := e.(*User)
		if !ok {
			return errors.New("expected User")
		}
		return fn(user)
	})
}

// UpdateGroup overwrites the Group in the store
func (a *BadgerAuthenticator) UpdateGroup(name string, fn func(*Group) error) error {
	g := Group{Name: name}
	return a.updateEntry(&g, func(e Entry) error {
		group, ok := e.(*Group)
		if !ok {
			return errors.New("expected Group")
		}
		return fn(group)
	})
}

// DeleteUser removes the User from the store.
// TODO: how to handle shadow fs
func (a *BadgerAuthenticator) DeleteUser(name string) error {
	return errors.New("stub")
}

// DeleteGroup removes the Group from the store and removes it from
// any Users.
// TODO: how to handle shadow fs
func (a *BadgerAuthenticator) DeleteGroup(name string) error {

	// get users
	users, err := a.GetUsers()
	if err != nil {
		return err
	}

	var changed []*User

	for _, u := range users {
		if _, ok := u.Groups[name]; ok {
			delete(u.Groups, name)
			changed = append(changed, u)
		}
		if u.PrimaryGroup == name {
			u.PrimaryGroup = ""
		}
	}

	err = a.db.Update(func(tx *badger.Txn) error {
		for _, u := range changed {
			if err := a.encodeAndUpdate(tx, u); err != nil {
				return err
			}
		}

		return tx.Delete([]byte("groups:" + name))
	})
	if err != nil {
		return err
	}

	return nil
}

// CheckPassword checks to see if the password is the correct one for
// the user. Any failure (i.e. user doesn't exist) returns false.
func (a *BadgerAuthenticator) CheckPassword(name, pass string) bool {
	u, err := a.GetUser(name)
	if err != nil {
		return false
	}

	if err := bcrypt.CompareHashAndPassword(u.Password, []byte(pass)); err != nil {
		return false
	}

	return true
}

// ChangePassword changes the password for the User
func (a *BadgerAuthenticator) ChangePassword(user, pass string) error {
	return errors.New("stub")
}

// CheckIPMask checks that the user ia authorised on the connecting ip / port
func (a *BadgerAuthenticator) CheckIP(name string, laddr, raddr net.Addr) bool {
	u, err := a.GetUser(name)
	if err != nil {
		// all these instances of just returning false might warrent an err
		// even if its just a log
		return false
	}

	// parse addresses
	_, lportStr, err := net.SplitHostPort(laddr.String())
	if err != nil {
		return false
	}

	lport, err := strconv.Atoi(lportStr)
	if err != nil {
		return false
	}

	host, rportStr, err := net.SplitHostPort(raddr.String())
	if err != nil {
		return false
	}

	rport, err := strconv.Atoi(rportStr)
	if err != nil {
		return false
	}

	// check all masks with a '*' to save us doing an ident lookup
	for idx := range u.IPMasks {
		parts := strings.Split(u.IPMasks[idx], "@")
		if len(parts) != 2 {
			continue
		}

		if parts[0] != "*" {
			continue
		}

		// bit inefficient, but im sure we will survive. can optimise later TM
		m, err := glob.Compile(parts[1], '.')
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"ERROR compiling mask %d for user %s\n",
				idx, u.Name,
			)
			continue
		}

		if m.Match(host) {
			return true
		}
	}

	ident, err := ident.Query(host, lport, rport, 10)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"ERROR querying ident for %s:%d from :%d: %s\n",
			host, rport, lport, err,
		)
		return false
	}

	identifier := strings.ToLower(ident.Identifier)

	for idx := range u.IPMasks {
		parts := strings.Split(u.IPMasks[idx], "@")
		if len(parts) != 2 {
			continue
		}
		if strings.ToLower(parts[0]) != identifier || parts[0] == "*" {
			continue
		}
		// bit inefficient, but im sure we will survive. can optimise later TM
		m, err := glob.Compile(parts[1], '.')
		if err != nil {
			fmt.Fprintf(
				os.Stderr,
				"ERROR compiling mask %d for user %s\n",
				idx, u.Name,
			)
			continue
		}

		if m.Match(host) {
			return true
		}
	}

	return false
}
