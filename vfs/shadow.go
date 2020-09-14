package vfs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v2"
	"github.com/pkg/errors"
	"github.com/segmentio/fasthash/fnv1a"
)

// Error thrown when the requested path does not exist
var ErrNoPath = errors.New("path does not exist")

// "constant" to the splitter used
var shadowEntrySplitter = []byte(":")

// Shadow represents a shadow filesystem where meta data is
// stored
type Shadow interface {
	Hash(string) []byte
	Add(string, string, string) error
	Get(string) (string, string, error)
	Remove(string) error
	Close() error
}

// ShadowStore uses an underlying badger key store value
// database to hold information about the filesystem.
// Paths are lower cased and hashed for security. And currently
// only
type ShadowStore struct {
	store *badger.DB
}

func NewShadowStore(db *badger.DB) Shadow {
	s := ShadowStore{
		store: db,
	}

	return &s
}

// Hash the given path into a byte using fnv1a
func (s *ShadowStore) Hash(path string) []byte {
	h := fnv1a.HashString64(strings.ToLower(path))

	// encode to bytes
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, h)

	return b
}

// Add a path with it's meta data to the store. Overwrites any
// existing value.
func (s *ShadowStore) Add(path, username, group string) error {
	key := s.Hash(path)
	val := []byte(strings.ToLower(fmt.Sprintf("%s:%s", username, group)))

	err := s.store.Update(func(txn *badger.Txn) error {
		if err := txn.Set(key, val); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// Get tries to retrieve the username and group for a path
func (s *ShadowStore) Get(path string) (string, string, error) {
	key := s.Hash(path)

	var user, group string

	err := s.store.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			parts := bytes.Split(val, shadowEntrySplitter)
			if len(parts) != 2 {
				return errors.Errorf("expected 2 parts to key: '%x': '%s'", key, string(val))
			}

			user = string(parts[0])
			group = string(parts[1])

			return nil
		})

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return "", "", ErrNoPath
		}

		// err has been set
		return "", "", err
	}

	return user, group, nil
}

// Remove deletes an entry
func (s *ShadowStore) Remove(path string) error {
	key := s.Hash(path)

	err := s.store.Update(func(txn *badger.Txn) error {
		if err := txn.Delete(key); err != nil {
			return err
		}

		return nil
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
