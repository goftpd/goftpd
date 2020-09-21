package acl

import (
	"github.com/dgraph-io/badger/v2"
	"github.com/pkg/errors"
)

type Authenticator interface {
	// create
	AddUser(string, string) (*User, error)
	AddGroup(string) error

	// get
	GetUser(string) (*User, error)
	GetGroup(string) (*Group, error)

	// save
	SaveUser(*User) error
	SaveGroup(*Group) error

	// delete
	DeleteUser(user string) error
	DeleteGroup(group string) error

	// utilities
	CheckPassword(string, string) bool
	ChangePassword(string, string) error
}

// BadgerAuthenticator implements an Authenticator using a badge key/value store
type BadgerAuthenticator struct {
	db *badger.DB
}

// AddUser creates a user setting the password
func (a *BadgerAuthenticator) AddUser(name, pass string) (*User, error) {
	return nil, errors.New("stub")
}

// AddGroup creates a Group
func (a *BadgerAuthenticator) AddGroup(name string) (*Group, error) {
	return nil, errors.New("stub")
}

// GetUser attempts to retrieve a User from the store using the name
func (a *BadgerAuthenticator) GetUser(name string) (*User, error) {
	return nil, errors.New("stub")
}

// GetGroup attempts to retrieve a Group from the store using the name
func (a *BadgerAuthenticator) GetGroup(name string) (*Group, error) {
	return nil, errors.New("stub")
}

// SaveUser overwrites the User in the store
func (a *BadgerAuthenticator) SaveUser(user *User) error {
	return errors.New("stub")
}

// SaveGroup overwrites the Group in the store
func (a *BadgerAuthenticator) SaveGroup(user *Group) error {
	return errors.New("stub")
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
	return errors.New("stub")
}

// CheckPassword checks to see if the password is the correct one for
// the user. Any failure (i.e. user doesn't exist) returns false.
func (a *BadgerAuthenticator) CheckPassword(user, pass string) bool {
	return false
}

// ChangePassword changes the password for the User
func (a *BadgerAuthenticator) ChangePassword(user, pass string) error {
	return errors.New("stub")
}
