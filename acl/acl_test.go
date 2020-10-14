package acl

import "testing"

func TestACLNewFromString(t *testing.T) {
	type test struct {
		line string
		want error
	}

	tests := map[string]test{
		"blank":                    test{"", ErrACLBadInput},
		"malformed negative":       test{"!", ErrACLMalformedNegative},
		"malformed user":           test{"-", ErrACLMalformedUser},
		"malformed negative user":  test{"!-", ErrACLMalformedUser},
		"bad user":                 test{"-*", ErrACLBadUser},
		"bad user characters":      test{"-@#$", ErrACLInvalidCharacters},
		"malformed group":          test{"=", ErrACLMalformedGroup},
		"malformed negative group": test{"!=", ErrACLMalformedGroup},
		"bad group":                test{"=*", ErrACLBadGroup},
		"bad group characters":     test{"=@#$", ErrACLInvalidCharacters},
		"bad special keyword":      test{"keyword", ErrACLInvalidCharacters},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			acl, err := NewFromString(tc.line)
			if err != tc.want {
				t.Fatalf("epected %#v, got %#v", tc.want, err)
			}

			if acl != nil {
				t.Fatalf("expected nil, got %#v", acl)
			}
		})
	}
}

func TestACLGadmin(t *testing.T) {
	acl, err := NewFromString("gadmin")
	if err != nil {
		t.Fatalf("expected nil, got %#v", err)
		return
	}

	if acl.allowed.gadmin != true {
		t.Fatal("expected acl.allowed.gadmin to be true")
	}

	if acl.blocked.gadmin == true {
		t.Fatal("expected acl.blocked.gadmin to be false")
	}

	acl, err = NewFromString("!gadmin")
	if err != nil {
		t.Fatalf("expected nil, got %#v", err)
		return
	}

	if acl.allowed.gadmin == true {
		t.Fatal("expected acl.allowed.gadmin to be false")
	}

	if acl.blocked.gadmin != true {
		t.Fatal("expected acl.blocked.gadmin to be true")
	}
}

func TestACLSelf(t *testing.T) {
	acl, err := NewFromString("self")
	if err != nil {
		t.Fatalf("expected nil, got %#v", err)
		return
	}

	if acl.allowed.self != true {
		t.Fatal("expected acl.allowed.self to be true")
	}

	if acl.blocked.self == true {
		t.Fatal("expected acl.blocked.self to be false")
	}

	acl, err = NewFromString("!self")
	if err != nil {
		t.Fatalf("expected nil, got %#v", err)
		return
	}

	if acl.allowed.self == true {
		t.Fatal("expected acl.allowed.self to be false")
	}

	if acl.blocked.self != true {
		t.Fatal("expected acl.blocked.self to be true")
	}
}

func TestACLMatchTarget(t *testing.T) {
	type test struct {
		line   string
		caller *User
		target *User
		want   bool
	}

	tests := map[string]test{
		"caller is nil": test{
			line:   "self",
			caller: nil,
			target: newUser("alice", "users"),
			want:   false,
		},
		"target is nil": test{
			line:   "self",
			caller: newUser("alice", "users"),
			target: nil,
			want:   false,
		},
		"self allowed": test{
			line:   "self",
			caller: newUser("alice", "users"),
			target: newUser("alice", "users"),
			want:   true,
		},
		"self not allowed": test{
			line:   "!self",
			caller: newUser("alice", "users"),
			target: newUser("alice", "users"),
			want:   false,
		},
		"gadmin allowed": test{
			line:   "gadmin",
			caller: newGadmin("alice", "users"),
			target: newUser("bob", "users"),
			want:   true,
		},
		"gadmin blocked": test{
			line:   "!gadmin",
			caller: newGadmin("alice", "users"),
			target: newUser("bob", "users"),
			want:   false,
		},
		"not my gadmin": test{
			line:   "gadmin",
			caller: newGadmin("alice", "users"),
			target: newUser("bob", "bobbers"),
			want:   false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			acl, err := NewFromString(tc.line)
			if err != nil {
				t.Fatalf("expected nil, got %#v", err)
				return
			}

			result := acl.MatchTarget(tc.caller, tc.target)
			if result != tc.want {
				t.Fatalf("expected %t, got %t", tc.want, result)
			}
		})
	}
}

func TestACLMatchTargetGroup(t *testing.T) {
	type test struct {
		line   string
		caller *User
		target *Group
		want   bool
	}

	tests := map[string]test{
		"caller is nil": test{
			line:   "self",
			caller: nil,
			target: newGroup("users"),
			want:   false,
		},
		"target is nil": test{
			line:   "self",
			caller: newUser("alice", "users"),
			target: nil,
			want:   false,
		},
		"user not allowed": test{
			line:   "gadmin",
			caller: newUser("alice", "users"),
			target: newGroup("users"),
			want:   false,
		},
		"gadmin allowed": test{
			line:   "gadmin",
			caller: newGadmin("alice", "users"),
			target: newGroup("users"),
			want:   true,
		},
		"user allowed gadmin not allowed": test{
			line:   "!gadmin =users",
			caller: newUser("alice", "users"),
			target: newGroup("users"),
			want:   true,
		},
		"gadmin not allowed": test{
			line:   "!gadmin",
			caller: newGadmin("alice", "users"),
			target: newGroup("users"),
			want:   false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			acl, err := NewFromString(tc.line)
			if err != nil {
				t.Fatalf("expected nil, got %#v", err)
				return
			}

			result := acl.MatchTargetGroup(tc.caller, tc.target)
			if result != tc.want {
				t.Fatalf("expected %t, got %t", tc.want, result)
			}
		})
	}
}
