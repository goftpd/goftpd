package acl

import (
	"testing"

	"github.com/dgraph-io/badger/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
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

func TestAuthUser(t *testing.T) {
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

	// update
	err = auth.UpdateUser("alice", func(u *User) error {
		u.AddGroup("testers")
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil, got %#v", err)
	}

	// test bad update
	badErr := errors.New("hello")
	err = auth.UpdateUser("alice", func(u *User) error {
		// try and remove, then return an error
		// this update should then not happen
		u.RemoveGroup("testers")
		return badErr
	})
	if err != badErr {
		t.Fatalf("expected badErr, got %#v", err)
	}

	// check update was ok
	updated, err := auth.GetUser("alice")
	if err != nil {
		t.Fatalf("expected err, got %#v", err)
	}

	if !updated.UpdatedAt.After(user.UpdatedAt) {
		t.Fatalf("expected UpdatedAt to be after, got %v", updated.UpdatedAt)
	}

	if !updated.HasGroup("testers") {
		t.Fatal("expected to have group 'testers'")
	}

	// TODO: delete
}

func TestAuthUserUpdateDoesntExist(t *testing.T) {
	auth := newAuthenticator(t)

	err := auth.UpdateUser("glftpd", func(u *User) error {
		return nil
	})
	if err != ErrUserDoesntExist {
		t.Fatalf("expected ErrUserDoesntExist, got %#v", err)
	}
}

func TestAuthGroupUpdateDoesntExist(t *testing.T) {
	auth := newAuthenticator(t)

	err := auth.UpdateGroup("glftpd", func(u *Group) error {
		return nil
	})
	if err != ErrUserDoesntExist {
		t.Fatalf("expected ErrGroupDoesntExist, got %#v", err)
	}
}

func TestAuthUserConflict(t *testing.T) {
	// TODO:
	// try and reliably create race conditions that mean the retry kicks in
	// and then check that both changes have been applied to the final
	// GetUser
}

func TestAuthUserBadDecode(t *testing.T) {
	// TODO:
	// somehow inject some bad entries that cause decode errors to propogate
	// up
}

func TestAuthGroup(t *testing.T) {
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

	// update
	err = auth.UpdateGroup("users", func(g *Group) error {
		g.Slots = 10
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil, got %#v", err)
	}

	// test bad update
	badErr := errors.New("hello")
	err = auth.UpdateGroup("users", func(g *Group) error {
		// try and remove, then return an error
		// this update should then not happen
		g.Slots = 15
		return badErr
	})
	if err != badErr {
		t.Fatalf("expected badErr, got %#v", err)
	}

	// check update was ok
	updated, err := auth.GetGroup("users")
	if err != nil {
		t.Fatalf("expected err, got %#v", err)
	}

	if updated.Slots != 10 {
		t.Fatalf("expected 10, got %d", updated.Slots)
	}

	err = auth.DeleteGroup("users")
	if err != nil {
		t.Fatalf("expected nil, got %#v", err)
	}

	postDelete, err := auth.GetGroups()
	if err != nil {
		t.Fatalf("expected nil, got %#v", err)
	}

	if len(postDelete) != 0 {
		t.Fatalf("expected 0, got %d", len(postDelete))
	}

}
