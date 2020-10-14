package acl

import (
	"bytes"
	"testing"
	"time"
)

func newGroup(name string) *Group {
	return &Group{
		Name:      name,
		Users:     make(map[string]*UserGroupMeta, 0),
		CreatedAt: time.Now(),
	}
}

func TestGroup(t *testing.T) {
	group := newGroup("groups")

	if len(group.Users) != 0 {
		t.Fatal("expected Users to be 0")
	}

	if group.AddUser("admin", "newgroup") {
		t.Fatal("expected AddUser to be false as no slots")
	}

	// set to 2 so we can test duplicate groups
	group.Slots = 2

	if !group.AddUser("admin", "newgroup") {
		t.Fatal("expected AddUser to be true")
	}

	if !group.AddUser("admin", "newgroup") {
		t.Fatal("expected duplicate AddUser to be true")
	}

	// reset to 1 to prevent slot overflow
	group.Slots = 1

	if group.AddUser("admin", "anothergroup") {
		t.Fatal("expected AddUser to be false as not enough slots")
	}

	if len(group.Users) != 1 {
		t.Fatal("expected Users to be 1")
	}

	if group.RemoveUser("nogroup") {
		t.Fatal("expected RemoveUser to be false as group doesnt exist")
	}

	if !group.RemoveUser("newgroup") {
		t.Fatal("expected RemoveUser to be true")
	}

	if group.RemoveUser("newgroup") {
		t.Fatal("expected duplicate RemoveUser to be false")
	}

	if len(group.Users) != 0 {
		t.Fatal("expected Users to be 0")
	}
}

func TestGroupKey(t *testing.T) {
	group := newGroup("users")
	if bytes.Compare(group.Key(), []byte("groups:users")) != 0 {
		t.Fatalf("expected %s, got %s", string(group.Key()), "groups:users")
	}
}

func TestGroupUpdatedAt(t *testing.T) {
	group := newGroup("users")
	if !group.UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt to be zero")
	}
	group.SetUpdatedAt()
	if group.UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt not to be zero")
	}
}
