package acl

import (
	"fmt"
	"testing"

	"github.com/gobwas/glob"
	"github.com/pkg/errors"
)

func TestNewRule(t *testing.T) {
	var tests = []struct {
		input string
		rule  Rule
		err   error
	}{
		{
			"download /path/test/dir -user !*",
			Rule{
				"/path/test/dir",
				PermissionScopeDownload,
				glob.MustCompile("/path/test/dir"),
				&ACL{
					collection{false, []string{"user"}, nil},
					collection{true, nil, nil},
				},
			},
			nil,
		},
		{
			"download /path/test/dir !-user *",
			Rule{
				"/path/test/dir",
				PermissionScopeDownload,
				glob.MustCompile("/path/test/dir"),
				&ACL{
					collection{true, nil, nil},
					collection{false, []string{"user"}, nil},
				},
			},
			nil,
		},
		{
			"notexist /path/test/dir !-user *",
			Rule{},
			errors.New("unknown permission scope 'notexist'"),
		},
		{
			"bad",
			Rule{},
			errors.New("rule requires minimum of 3 fields"),
		},
		{
			"bad line",
			Rule{},
			errors.New("rule requires minimum of 3 fields"),
		},
		{
			"download /path/test !-*",
			Rule{
				"/path/test",
				PermissionScopeDownload,
				glob.MustCompile("/path/test"),
				nil,
			},
			errors.New("bad user '*'"),
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.input,
			func(t *testing.T) {
				rule, err := NewRule(tt.input)
				checkErr(t, err, tt.err)

				if tt.rule.path != rule.path {
					t.Errorf("expected path to be '%s' but got '%s'", tt.rule.path, rule.path)
				}

				if tt.rule.scope != rule.scope {
					t.Errorf("expected scope to be '%s' but got '%s'", tt.rule.scope, rule.scope)
				}

				if tt.rule.acl != nil && rule.acl != nil {
					if !compareACL(tt.rule.acl, rule.acl) {
						t.Error("acl do not match")
					}
				}
			},
		)
	}
}

func TestNewPermissions(t *testing.T) {
	var tests = []struct {
		lines []string
		err   error
	}{
		{
			[]string{},
			nil,
		},
		{
			[]string{
				"download /dir/a *",
				"download /dir/b !*",
			},
			nil,
		},
	}

	for idx, tt := range tests {
		t.Run(
			fmt.Sprintf("%d", idx),
			func(t *testing.T) {

				var rules []Rule
				for _, l := range tt.lines {
					r, err := NewRule(l)
					if err != nil {
						t.Fatalf("unable to parse rule '%s': %s", l, err)
					}
					rules = append(rules, r)
				}

				_, err := NewPermissions(rules)
				checkErr(t, err, tt.err)
			},
		)
	}
}

func TestPermissionsCheck(t *testing.T) {
	var tests = []struct {
		input    string
		path     string
		scope    PermissionScope
		user     *User
		expected bool
	}{
		{
			"download /dir/a *",
			"/dir/a",
			PermissionScopeDownload,
			newTestUser("user"),
			true,
		},
		{
			"download /dir/a !*",
			"/dir/a",
			PermissionScopeDownload,
			newTestUser("user"),
			false,
		},
		{
			"download /dir/a -user !*",
			"/dir/a",
			PermissionScopeDownload,
			newTestUser("user"),
			true,
		},
		{
			"download /dir/a =group !*",
			"/dir/a",
			PermissionScopeDownload,
			newTestUser("user", "group"),
			true,
		},
		{
			"download /** =group !*",
			"/dir/a",
			PermissionScopeDownload,
			newTestUser("user", "group"),
			true,
		},
		{
			"download / =group !*",
			"/dir/a",
			PermissionScopeUpload,
			newTestUser("user", "group"),
			false,
		},
		{
			"download /some/path =group !*",
			"/dir/a",
			PermissionScopeDownload,
			newTestUser("user", "group"),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.input,
			func(t *testing.T) {
				r, err := NewRule(tt.input)
				if err != nil {
					t.Fatalf("unable to parse rule '%s': %s", tt.input, err)
				}

				p, err := NewPermissions([]Rule{r})
				if err != nil {
					t.Fatalf("unable to create Permissions: %s", err)
				}

				allowed := p.Match(tt.scope, tt.path, tt.user)
				if allowed != tt.expected {
					t.Errorf("expected %t got %t", tt.expected, allowed)
				}
			},
		)
	}
}
