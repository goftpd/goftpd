package vfs

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v2"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v5"
)

// Error thrown when the requested path does not exist
var ErrNoPath = errors.New("path does not exist")

// "constants" to the splitter used
var shadowEntrySplitter = ":"
var shadowEntrySplitterBytes = []byte(shadowEntrySplitter)

type Entry struct {
	IsDir bool

	// an id would be nicer for renaming
	User  string
	Group string

	CRC uint32

	// meta
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewEntry(user, group string) Entry {
	return Entry{
		User:      strings.ToLower(user),
		Group:     strings.ToLower(group),
		CreatedAt: time.Now(),
	}
}

// prints the crc in hex, or if nothing is set 00000000
func (e Entry) CRCHex() string {
	if e.CRC == 0 {
		return "00000000"
	}
	return fmt.Sprintf("%x", e.CRC)
}

// Shadow represents a shadow filesystem where meta data is
// stored
type Shadow interface {
	Set(string, *Entry) error
	Get(string) (*Entry, error)
	Remove(string) error
	Close() error
}

// ShadowStore uses an underlying badger key store value
// database to hold information about the filesystem.
// Paths are lower cased and hashed for security. And currently
// only
type ShadowStore struct {
	store      *badger.DB
	bufferPool sync.Pool
}

// NewShadowStore creates a new ShadowStore with the given badger db. Caller
// is responsible for opening the db to make it easier to test.
func NewShadowStore(db *badger.DB) *ShadowStore {
	s := ShadowStore{
		store: db,
		bufferPool: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}

	return &s
}

// Set a path with it's meta data to the store. Overwrites any
// existing value.
func (s *ShadowStore) Set(path string, entry *Entry) error {
	key := []byte(strings.ToLower(path))

	entry.UpdatedAt = time.Now()

	err := s.store.Update(func(tx *badger.Txn) error {
		enc := msgpack.GetEncoder()
		defer msgpack.PutEncoder(enc)

		b := s.bufferPool.Get().(*bytes.Buffer)
		b.Reset()
		defer s.bufferPool.Put(b)

		enc.Reset(b)

		if err := enc.Encode(entry); err != nil {
			return err
		}

		return tx.Set(key, b.Bytes())
	})

	if err != nil {
		return err
	}

	return nil
}

// Get tries to retrieve the user and group for a path
func (s *ShadowStore) Get(path string) (*Entry, error) {
	key := []byte(strings.ToLower(path))

	var entry Entry

	err := s.store.View(func(tx *badger.Txn) error {
		item, err := tx.Get(key)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			dec := msgpack.GetDecoder()
			defer msgpack.PutDecoder(dec)

			dec.ResetBytes(val)

			return dec.Decode(&entry)
		})
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrNoPath
		}

		// err has been set
		return nil, err
	}

	return &entry, nil
}

// Remove deletes an entry from the store
func (s *ShadowStore) Remove(path string) error {
	key := []byte(strings.ToLower(path))

	err := s.store.Update(func(tx *badger.Txn) error {
		return tx.Delete(key)
	})

	if err != nil {
		return err
	}

	return nil
}

// Close closes the underlying badger store
func (s *ShadowStore) Close() error {
	return s.store.Close()
}
