package acl

import (
	"fmt"
	"testing"
)

func TestNewFromString(t *testing.T) {
	var tests = []struct {
		input string
		err   string
	}{
		{"-user =group 1", ""},
		{"", "no input string given"},
		{"- =group 1", "expected string after '-'"},
		{"-user = 1", "expected string after '='"},
		{"! -user =group !flag", "expected string after '!'"},
		{"1 2 3 4 -u1 -u2 -u3 !-u4 =g1", ""},
	}

	for _, tt := range tests {
		t.Run(
			tt.input,
			func(t *testing.T) {
				_, err := NewFromString(tt.input)

				if err != nil && len(tt.err) == 0 {
					t.Errorf("expected nil but got: '%s'", err)
					return
				}

				if err != nil && tt.err != err.Error() {
					t.Errorf("expected '%s' but got: '%s'", tt.err, err)
					return
				}

				if err == nil && len(tt.err) > 0 {
					t.Errorf("expected '%s' but got nil", err)
					return
				}
			},
		)
	}
}

type TestUser struct {
	name   string
	groups []string
	flags  []string
}

func (u TestUser) Name() string {
	return u.name
}

func (u TestUser) Groups() []string {
	return u.groups
}

func (u TestUser) Flags() []string {
	return u.flags
}

func TestAllowed(t *testing.T) {
	var tests = []struct {
		input    string
		user     TestUser
		expected bool
	}{
		{
			"-testUser *",
			TestUser{"testUser", []string{""}, []string{""}},
			true,
		},
		// check specifying user overrides blocking all
		{
			"-testUser !*",
			TestUser{"testUser", []string{""}, []string{""}},
			true,
		},
		// check denying user overrides allowing all
		{
			"!-testUser *",
			TestUser{"testUser", []string{""}, []string{""}},
			false,
		},
		// check denying user overrides allowing group
		{
			"!-testUser =testGroup",
			TestUser{"testUser", []string{"testGroup"}, []string{""}},
			false,
		},
		// check denying user overrides flag
		{
			"!-testUser 1",
			TestUser{"testUser", []string{""}, []string{"1"}},
			false,
		},
		// check capitalisation doesnt matter
		{
			"-testUser !*",
			TestUser{"testuser", []string{}, []string{""}},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.input,
			func(t *testing.T) {
				acl, err := NewFromString(tt.input)
				if err != nil {
					t.Errorf("expected nil but got: '%s'", err)
					return
				}

				if acl.Allowed(tt.user) != tt.expected {
					fmt.Printf("acl: %+v\ntt: %+v\n", acl, tt)
					t.Errorf("expected %t but got: %t", tt.expected, !tt.expected)
					return
				}
			},
		)
	}
}
