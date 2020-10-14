package acl

import (
	"bytes"
	"testing"
)

var DummyUser = &User{}

func newUser(name string, groups ...string) *User {
	user := &User{
		Name:   name,
		Groups: make(map[string]*GroupSettings, len(groups)),
	}

	for idx, g := range groups {
		if idx == 0 {
			user.PrimaryGroup = g
		}

		user.AddGroup(g)
	}

	return user
}

func newGadmin(name string, groups ...string) *User {
	user := newUser(name, groups...)
	for g := range user.Groups {
		user.Groups[g].IsAdmin = true
	}
	return user
}

func TestSuperUser(t *testing.T) {
	user := newUser("alice", "users")

	if user == SuperUser {
		t.Fatalf("user should not match SuperUser")
	}

	if DummyUser == SuperUser {
		t.Fatalf("DummyUser should not match SuperUser")
	}
}

func TestUserHasGroup(t *testing.T) {
	user := newUser("alice", "USERS")

	if !user.HasGroup("users") {
		t.Fatal("expected user to have group")
	}

	if !user.HasGroup("UsErS") {
		t.Fatal("expected user to have group (case insensitive)")
	}

	if user.HasGroup("agroup") {
		t.Fatal("unexpected group found")
	}
}

func TestUserAddGroup(t *testing.T) {
	user := newUser("alice")

	user.AddGroup("users")
	if !user.HasGroup("users") {
		t.Fatal("expected user to have group")
	}

	user.AddGroup("users")
	if !user.HasGroup("users") {
		t.Fatal("expected user to have group")
	}
}

func TestUserRemoveGroup(t *testing.T) {
	user := newUser("alice", "USERS")

	if !user.HasGroup("users") {
		t.Fatal("expected user to have group")
	}

	user.RemoveGroup("uSeRs")
	// dupe shuld be noop
	user.RemoveGroup("uSeRs")

	if user.HasGroup("!agroup") {
		t.Fatal("expected user to not have group")
	}

	if user.PrimaryGroup == "users" {
		t.Fatal("expected user to have no PrimaryGroup")
	}
}

func TestUserDeleteReadd(t *testing.T) {
	user := newUser("alice", "USERS")
	if !user.DeletedAt.IsZero() {
		t.Fatal("expected user to have zero DeletedAt")
	}

	user.Delete()

	if user.DeletedAt.IsZero() {
		t.Fatal("expected user to have set DeletedAt")
	}

	user.Readd()

	if !user.DeletedAt.IsZero() {
		t.Fatal("expected user to have zero DeletedAt")
	}
}

func TestUserAddIP(t *testing.T) {
	type test struct {
		input string
		want  error
	}

	newTest := func(input string, want error) test {
		return test{
			input: input,
			want:  want,
		}
	}

	tests := map[string]test{
		"malformed mask":  newTest("*", ErrUserIPMalformed),
		"required octets": newTest("*@*", ErrUserIPRequiredOctets),
		"bad ident glob":  newTest("[*@1.2.3.*", ErrUserIPBadGlob),
		"bad octet glob":  newTest("*@1.2.3.[*", ErrUserIPBadGlob),
		"ok ip":           newTest("*@1.2.3.4", nil),
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			user := newUser("alice", "users")
			err := user.AddIP(tc.input)
			if err != tc.want {
				t.Fatalf("expected: %v, got: %v", tc.want, err)
			}
		})
	}

	// special cases
	func() {
		user := newUser("alice", "users")
		if err := user.AddIP("*@1.2.3.4"); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}

		if err := user.AddIP("*@1.2.3.4"); err != ErrUserIPExists {
			t.Fatalf("expected %v, got %v", ErrUserIPExists, err)
		}
	}()
}

func TestUserDeleteIP(t *testing.T) {
	user := newUser("alice", "users")
	if err := user.AddIP("*@1.2.3.4"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	if user.DeleteIP("ident@5.6.7.8") {
		t.Fatal("expected false, got true")
	}

	if !user.DeleteIP("*@1.2.3.4") {
		t.Fatal("expected true, got false")
	}

	if len(user.IPMasks) != 0 {
		t.Fatalf("expected 0, got %d", len(user.IPMasks))
	}
}

func TestUserKey(t *testing.T) {
	user := newUser("aLiCe", "users")
	if bytes.Compare(user.Key(), []byte("users:alice")) != 0 {
		t.Fatalf("expected 'users:alice', got '%v'", string(user.Key()))
	}
}

func TestUserUpdatedAt(t *testing.T) {
	user := newUser("aLiCe", "users")
	if !user.UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt to be zero")
	}

	user.SetUpdatedAt()

	if user.UpdatedAt.IsZero() {
		t.Fatal("expected UpdatedAt to not be zero")
	}

}
