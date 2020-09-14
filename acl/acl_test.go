package acl

import (
	"testing"
)

func TestNewFromString(t *testing.T) {
	var tests = []struct {
		input string
		err   string
	}{
		{"-user =group", ""},
		{"", "no input string given"},
		{"- =group", "expected string after '-'"},
		{"-user =", "expected string after '='"},
		{"! -user =group", "expected string after '!'"},
		{"-u1 -u2 -u3 !-u4 =g1", ""},
		{"something", "unexpected string in acl input: 'something'"},
		{"-*", "bad user '*'"},
		{"=*", "bad group '*'"},
		{"-_", "user contains invalid characters: '_'"},
		{"=:", "group contains invalid characters: ':'"},
	}

	for _, tt := range tests {
		t.Run(
			tt.input,
			func(t *testing.T) {
				_, err := NewFromString(tt.input)

				if err != nil && len(tt.err) == 0 {
					t.Fatalf("expected nil but got: '%s'", err)
				}

				if err != nil && tt.err != err.Error() {
					t.Fatalf("expected '%s' but got: '%s'", tt.err, err)
				}

				if err == nil && len(tt.err) > 0 {
					t.Fatalf("expected '%s' but got nil", err)
				}
			},
		)
	}
}

type TestUser struct {
	name   string
	groups []string
}

func (u TestUser) Name() string {
	return u.name
}

func (u TestUser) Groups() []string {
	return u.groups
}

func (u TestUser) PrimaryGroup() string {
	if len(u.groups) > 0 {
		return u.groups[0]
	}
	return "nobody"
}

func TestAllowed(t *testing.T) {
	var tests = []struct {
		input    string
		user     TestUser
		expected bool
	}{
		{
			"-testUser *",
			TestUser{"testUser", nil},
			true,
		},
		// check specifying user overrides blocking all
		{
			"-testUser !*",
			TestUser{"testUser", nil},
			true,
		},
		// check denying user overrides allowing all
		{
			"!-testUser *",
			TestUser{"testUser", nil},
			false,
		},
		// check denying user overrides allowing group
		{
			"!-testUser =testGroup",
			TestUser{"testUser", []string{"testGroup"}},
			false,
		},
		// check capitalisation doesnt matter
		{
			"-testUser !*",
			TestUser{"testuser", nil},
			true,
		},
		// check banned group
		{
			"!=testGroup *",
			TestUser{"testuser", []string{"testGroup"}},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.input,
			func(t *testing.T) {
				acl, err := NewFromString(tt.input)
				if err != nil {
					t.Fatalf("expected nil but got: '%s'", err)
				}

				if acl.Allowed(tt.user) != tt.expected {
					t.Errorf("expected %t but got: %t", tt.expected, !tt.expected)
				}
			},
		)
	}
}
