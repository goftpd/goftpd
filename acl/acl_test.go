package acl

import (
	"testing"

	"github.com/pkg/errors"
)

func TestNewFromString(t *testing.T) {
	var tests = []struct {
		input string
		err   error
	}{
		{
			"-user =group",
			nil,
		},
		{
			"",
			errors.New("no input string given"),
		},
		{
			"- =group",
			errors.New("expected string after '-'"),
		},
		{
			"-user =",
			errors.New("expected string after '='"),
		},
		{
			"! -user =group",
			errors.New("expected string after '!'"),
		},
		{
			"-u1 -u2 -u3 !-u4 =g1",
			nil,
		},
		{
			"something",
			errors.New("unexpected string in acl input: 'something'"),
		},
		{
			"-*",
			errors.New("bad user '*'"),
		},
		{
			"=*",
			errors.New("bad group '*'"),
		},
		{
			"-_",
			errors.New("user contains invalid characters: '_'"),
		},
		{
			"=:",
			errors.New("group contains invalid characters: ':'"),
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.input,
			func(t *testing.T) {
				_, err := NewFromString(tt.input)
				checkErr(t, err, tt.err)
			},
		)
	}
}

func TestAllowed(t *testing.T) {
	var tests = []struct {
		input    string
		user     *User
		expected bool
	}{
		{
			"-testUser *",
			newTestUser("testUser"),
			true,
		},
		// check specifying user overrides blocking all
		{
			"-testUser !*",
			newTestUser("testUser"),
			true,
		},
		// check denying user overrides allowing all
		{
			"!-testUser *",
			newTestUser("testUser"),
			false,
		},
		// check denying user overrides allowing group
		{
			"!-testUser =testGroup",
			newTestUser("testUser", "testGroup"),
			false,
		},
		// check capitalisation doesnt matter
		{
			"-testUser !*",
			newTestUser("testUser"),
			true,
		},
		// check banned group
		{
			"!=testGroup *",
			newTestUser("testUser", "testGroup"),
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

				if acl.Match(tt.user) != tt.expected {
					t.Errorf("expected %t but got: %t", tt.expected, !tt.expected)
				}
			},
		)
	}
}
