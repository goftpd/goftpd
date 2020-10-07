package vfs

import (
	"strings"
	"testing"

	"github.com/dgraph-io/badger/v2"
)

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
		entry := NewEntry(e.user, e.group)
		if err := ss.Set(e.path, &entry); err != nil {
			t.Fatalf("unexpected err adding %s:%s:%s: %s", e.path, e.user, e.group, err)
		}
	}

	for _, e := range expected {
		entry, err := ss.Get(e.path)
		if err != nil {
			t.Fatalf("unexpected err getting %s: %s", e.path, err)
		}

		if entry.User != e.user {
			t.Errorf("expected user for '%s' to be '%s' got '%s'", e.path, e.user, entry.User)
		}

		if entry.Group != e.group {
			t.Errorf("expected group for '%s' to be '%s' got '%s'", e.path, e.group, entry.Group)
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

	entry := NewEntry(user, group)
	if err := ss.Set(path, &entry); err != nil {
		t.Fatalf("expected nil on add got: %s", err)
	}

	gentry, err := ss.Get(path)
	if err != nil {
		t.Fatalf("expected nil on get got: %s", err)
	}

	if gentry.User != user {
		t.Errorf("expected user to be '%s' got '%s'", user, gentry.User)
	}

	if gentry.Group != group {
		t.Errorf("expected group to be '%s' got '%s'", group, gentry.Group)
	}

	if err := ss.Remove(path); err != nil {
		t.Fatalf("expected nil on remove got: %s", err)
	}

	if _, err := ss.Get(path); err != ErrNoPath {
		t.Fatalf("expected ErrNoPath on post remove get, got: %s", err)
	}
}

func TestShadowStoreBadValue(t *testing.T) {
	ss := newMemoryShadowStore(t)
	defer closeMemoryShadowStore(t, ss)

	path := "/a/b/c"
	badValue := "bad"

	key := []byte(strings.ToLower(path))

	err := ss.(*ShadowStore).store.Update(func(txn *badger.Txn) error {
		return txn.Set(key, []byte(badValue))
	})
	if err != nil {
		t.Fatalf("unexpected error for manual insert of key: %s", err)
	}

	_, err = ss.Get(path)
	if err == nil {
		t.Fatal("expected error on get")
	}

	if !strings.Contains(err.Error(), "msgpack: invalid") {
		t.Fatalf("unexpected error: %s", err)
	}
}
