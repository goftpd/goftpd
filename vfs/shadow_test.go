package vfs

import (
	"fmt"
	"testing"

	"github.com/dgraph-io/badger/v2"
)

func newMemoryShadowStore(t *testing.T) Shadow {
	opt := badger.DefaultOptions("").WithInMemory(true)

	db, err := badger.Open(opt)
	if err != nil {
		t.Errorf("error opening db: %s", err)
		return nil
	}

	ss := NewShadowStore(db)

	return ss
}

func closeMemoryShadowStore(t *testing.T, ss Shadow) {
	if err := ss.Close(); err != nil {
		t.Errorf("error closing shadow: %s", err)
	}
}

func TestShadowStore(t *testing.T) {
	var entries = []struct {
		path  string
		user  string
		group string
	}{
		{"/a", "user0", "group0"},
		{"/a/b", "user1", "group1"},
		{"/ab/file.jpg", "user2", "group2"},
		{"/A/B", "user3", "GROUP3"},
	}

	var expected = []struct {
		path  string
		user  string
		group string
	}{
		{"/a", "user0", "group0"},
		{"/ab/file.jpg", "user2", "group2"},
		{"/a/b", "user3", "group3"},
	}

	ss := newMemoryShadowStore(t)
	defer closeMemoryShadowStore(t, ss)

	// do adds
	for _, e := range entries {
		if err := ss.Add(e.path, e.user, e.group); err != nil {
			t.Errorf("unexpected err adding %s:%s:%s: %s", e.path, e.user, e.group, err)
			return
		}
	}

	for _, e := range expected {
		user, group, err := ss.Get(e.path)
		if err != nil {
			t.Errorf("unexpected err getting %s: %s", e.path, err)
			return
		}

		if user != e.user {
			t.Errorf("expected user for '%s' to be '%s' got '%s'", e.path, e.user, user)
			return
		}

		if group != e.group {
			t.Errorf("expected group for '%s' to be '%s' got '%s'", e.path, e.group, group)
			return
		}
	}
}

func TestShadowStoreRemove(t *testing.T) {
	ss := newMemoryShadowStore(t)
	defer closeMemoryShadowStore(t, ss)

	var (
		path  string = "/a/b/c"
		user  string = "user"
		group string = "group"
	)

	if err := ss.Add(path, user, group); err != nil {
		t.Errorf("expected nil on add got: %s", err)
		return
	}

	guser, ggroup, err := ss.Get(path)
	if err != nil {
		t.Errorf("expected nil on get got: %s", err)
		return
	}

	if guser != user {
		t.Errorf("expected user to be '%s' got '%s'", user, guser)
		return
	}

	if ggroup != group {
		t.Errorf("expected group to be '%s' got '%s'", group, ggroup)
		return
	}

	if err := ss.Remove(path); err != nil {
		t.Errorf("expected nil on remove got: %s", err)
		return
	}

	if _, _, err := ss.Get(path); err != ErrNoPath {
		t.Errorf("expected ErrNoPath on post remove get, got: %s", err)
		return
	}
}

func TestShadowStoreCreateVal(t *testing.T) {
	var ss ShadowStore

	val, err := ss.createVal("user", "group")
	if err != nil {
		t.Errorf("unexpected err: %s", err)
		return
	}

	if string(val) != "user:group" {
		t.Errorf("unexpected val: '%s'", string(val))
		return
	}

	val, err = ss.createVal("user:", "group")
	if err == nil {
		t.Error("expected bad user createVal to error")
		return
	}

	if err.Error() != "user can't contain ':'" {
		t.Errorf("unexpected error for bad user createVal: %s", err)
		return
	}

	val, err = ss.createVal("user", "group:")
	if err == nil {
		t.Error("expected bad group createVal to error")
		return
	}

	if err.Error() != "group can't contain ':'" {
		t.Errorf("unexpected error for bad group createVal: %s", err)
		return
	}
}

func TestShadowStoreBadValue(t *testing.T) {
	ss := newMemoryShadowStore(t)
	defer closeMemoryShadowStore(t, ss)

	path := "/a/b/c"
	badValue := "bad"

	key := ss.Hash(path)

	err := ss.(*ShadowStore).store.Update(func(txn *badger.Txn) error {
		return txn.Set(key, []byte(badValue))
	})
	if err != nil {
		t.Errorf("unexpected error for manual insert of key: %s", err)
		return
	}

	_, _, err = ss.Get(path)
	if err == nil {
		t.Error("expected error on get")
		return
	}

	expectedErr := fmt.Sprintf("expected 2 parts to key: '%x': '%s'", key, badValue)

	if err.Error() != expectedErr {
		t.Errorf("unexpected error: %s", err)
		return
	}
}
