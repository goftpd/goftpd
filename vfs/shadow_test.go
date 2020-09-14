package vfs

import (
	"fmt"
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
		if err := ss.Set(e.path, e.user, e.group); err != nil {
			t.Fatalf("unexpected err adding %s:%s:%s: %s", e.path, e.user, e.group, err)
		}
	}

	for _, e := range expected {
		user, group, err := ss.Get(e.path)
		if err != nil {
			t.Fatalf("unexpected err getting %s: %s", e.path, err)
		}

		if user != e.user {
			t.Errorf("expected user for '%s' to be '%s' got '%s'", e.path, e.user, user)
		}

		if group != e.group {
			t.Errorf("expected group for '%s' to be '%s' got '%s'", e.path, e.group, group)
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

	if err := ss.Set(path, user, group); err != nil {
		t.Fatalf("expected nil on add got: %s", err)
	}

	guser, ggroup, err := ss.Get(path)
	if err != nil {
		t.Fatalf("expected nil on get got: %s", err)
	}

	if guser != user {
		t.Errorf("expected user to be '%s' got '%s'", user, guser)
	}

	if ggroup != group {
		t.Errorf("expected group to be '%s' got '%s'", group, ggroup)
	}

	if err := ss.Remove(path); err != nil {
		t.Fatalf("expected nil on remove got: %s", err)
	}

	if _, _, err := ss.Get(path); err != ErrNoPath {
		t.Fatalf("expected ErrNoPath on post remove get, got: %s", err)
	}
}

func TestShadowStoreCreateVal(t *testing.T) {
	var ss ShadowStore

	val, err := ss.createVal("user", "group")
	if err != nil {
		t.Fatalf("unexpected err: %s", err)
	}

	if string(val) != "user:group" {
		t.Fatalf("unexpected val: '%s'", string(val))
	}

	val, err = ss.createVal("user:", "group")
	if err == nil {
		t.Fatal("expected bad user createVal to error")
	}

	if err.Error() != "user can't contain ':'" {
		t.Fatalf("unexpected error for bad user createVal: %s", err)
	}

	val, err = ss.createVal("user", "group:")
	if err == nil {
		t.Fatal("expected bad group createVal to error")
	}

	if err.Error() != "group can't contain ':'" {
		t.Fatalf("unexpected error for bad group createVal: %s", err)
	}

	err = ss.Set("/", "user:", "group")
	if err == nil {
		t.Fatal("expected err but got nil")
	}
	if err.Error() != "user can't contain ':'" {
		t.Fatalf("unexpected error for bad user Set: %s", err)
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
		t.Fatalf("unexpected error for manual insert of key: %s", err)
	}

	_, _, err = ss.Get(path)
	if err == nil {
		t.Fatal("expected error on get")
	}

	expectedErr := fmt.Sprintf("expected 2 parts to key: '%x': '%s'", key, badValue)

	if err.Error() != expectedErr {
		t.Fatalf("unexpected error: %s", err)
	}
}
