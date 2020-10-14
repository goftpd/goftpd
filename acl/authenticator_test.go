package acl

import (
	"testing"

	"github.com/dgraph-io/badger/v2"
	"github.com/google/go-cmp/cmp"
)

func newAuthenticator(t *testing.T) Authenticator {
	opt := badger.DefaultOptions("").WithInMemory(true)
	opt.Logger = nil

	db, err := badger.Open(opt)
	if err != nil {
		t.Fatalf("error opening db: %s", err)
	}

	return NewBadgerAuthenticator(db)
}

func TestUser(t *testing.T) {
	auth := newAuthenticator(t)

	// try get none existent user
	_, err := auth.GetUser("alice")
	if err != ErrUserDoesntExist {
		t.Fatalf("expected %#v, got %#v", ErrUserDoesntExist, err)
	}

	users, err := auth.GetUsers()
	if err != nil {
		t.Fatalf("expected nil, got %#v", err)
	}

	if len(users) != 0 {
		t.Fatalf("expected 0, got %d", len(users))
	}

	// create user
	user, err := auth.AddUser("alice", "manygoodpasswords")
	if err != nil {
		t.Fatalf("expected nil, got %#v", err)
	}

	// try recreate user get: ErrUserExists
	_, err = auth.AddUser("alice", "manygoodpasswords")
	if err != ErrUserExists {
		t.Fatalf("expected %#v, got %#v", ErrUserExists, err)
	}

	// get user
	got, err := auth.GetUser("alice")
	if err != nil {
		t.Fatalf("expected nil, got %#v", err)
	}

	// cmp
	diff := cmp.Diff(user, got)
	if len(diff) > 0 {
		t.Fatal(diff)
	}

	users, err = auth.GetUsers()
	if err != nil {
		t.Fatalf("expected nil, got %#v", err)
	}

	if len(users) != 1 {
		t.Fatalf("expected 1, got %d", len(users))
	}

	diff = cmp.Diff(user, users[0])
	if len(diff) > 0 {
		t.Fatal(diff)
	}

	// TODO: delete
}

func TestAddGroup(t *testing.T) {
	auth := newAuthenticator(t)

	groups, err := auth.GetGroups()
	if err != nil {
		t.Fatalf("expected nil, got %#v", err)
	}

	if len(groups) != 0 {
		t.Fatalf("expected 0, got %d", len(groups))
	}

	// create group
	group, err := auth.AddGroup("users")
	if err != nil {
		t.Fatalf("expected nil, got %#v", err)
	}

	// try recreate group get: ErrGroupExists
	_, err = auth.AddGroup("users")
	if err != ErrGroupExists {
		t.Fatalf("expected %#v, got %#v", ErrGroupExists, err)
	}

	// get group
	got, err := auth.GetGroup("users")
	if err != nil {
		t.Fatalf("expected nil, got %#v", err)
	}

	// cmp
	diff := cmp.Diff(group, got)
	if len(diff) > 0 {
		t.Fatal(diff)
	}

	groups, err = auth.GetGroups()
	if err != nil {
		t.Fatalf("expected nil, got %#v", err)
	}

	if len(groups) != 1 {
		t.Fatalf("expected 1, got %d", len(groups))
	}

	diff = cmp.Diff(group, groups[0])
	if len(diff) > 0 {
		t.Fatal(diff)
	}
}
