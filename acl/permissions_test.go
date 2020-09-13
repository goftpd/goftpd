package acl

import (
	"fmt"
	"testing"
)

func compareACL(a, b *ACL) bool {
	if !compareSlices(a.allowed.users, b.allowed.users) {
		return false
	}

	if !compareSlices(a.allowed.groups, b.allowed.groups) {
		return false
	}

	return true
}

func compareSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for _, i := range a {
		var match bool
		for _, j := range b {
			if i == j {
				match = true
				break
			}
		}

		if !match {
			return false
		}
	}

	return true
}

func TestNewRule(t *testing.T) {
	var tests = []struct {
		input string
		rule  Rule
		err   string
	}{
		{
			"download /path/test/dir -user !*",
			Rule{
				"/path/test/dir",
				PermissionScopeDownload,
				&ACL{
					collection{false, []string{"user"}, nil},
					collection{true, nil, nil},
				},
			},
			"",
		},
		{
			"download /path/test/dir !-user *",
			Rule{
				"/path/test/dir",
				PermissionScopeDownload,
				&ACL{
					collection{true, nil, nil},
					collection{false, []string{"user"}, nil},
				},
			},
			"",
		},
		{
			"notexist /path/test/dir !-user *",
			Rule{},
			"unknown permission scope 'notexist'",
		},
		{
			"bad",
			Rule{},
			"rule requires minimum of 3 fields",
		},
		{
			"bad line",
			Rule{},
			"rule requires minimum of 3 fields",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.input,
			func(t *testing.T) {
				rule, err := NewRule(tt.input)
				if err != nil && len(tt.err) == 0 {
					t.Errorf("expected nil but got: '%s'", err)
					return
				}

				if err != nil && tt.err != err.Error() {
					t.Errorf("expected '%s' but got: '%s'", tt.err, err)
					return
				}

				if err == nil && len(tt.err) > 0 {
					t.Errorf("expected '%s' but got nil", tt.err)
					return
				}

				if tt.rule.path != rule.path {
					t.Errorf("expected path to be '%s' but got '%s'", tt.rule.path, rule.path)
					return
				}

				if tt.rule.scope != rule.scope {
					t.Errorf("expected scope to be '%s' but got '%s'", tt.rule.scope, rule.scope)
					return
				}

				if tt.rule.acl != nil && rule.acl != nil {
					if !compareACL(tt.rule.acl, rule.acl) {
						t.Error("acl do not match")
						return
					}
				}
			},
		)
	}
}

func TestNewPermissions(t *testing.T) {
	var tests = []struct {
		lines []string
		err   string
	}{
		{
			[]string{},
			"",
		},
		{
			[]string{
				"download /dir/a *",
				"download /dir/b !*",
			},
			"",
		},
		{
			[]string{
				"download /dir/a *",
				"download /dir/a !*",
			},
			"path '/dir/a' for scope 'download' already exists",
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
						t.Errorf("unable to parse rule '%s': %s", l, err)
						return
					}
					rules = append(rules, r)
				}
				_, err := NewPermissions(rules)
				if err != nil && len(tt.err) == 0 {
					t.Errorf("expected nil but got: '%s'", err)
					return
				}

				if err != nil && len(tt.err) > 0 && err.Error() != tt.err {
					t.Errorf("expected '%s' but got: '%s'", tt.err, err)
					return
				}

				if err == nil && len(tt.err) > 0 {
					t.Errorf("expected '%s' but got nil", tt.err)
					return
				}
			},
		)
	}
}

func TestPermissionsCheck(t *testing.T) {
	var tests = []struct {
		input    string
		user     TestUser
		expected bool
	}{
		{
			"download /dir/a *",
			TestUser{"user", nil},
			true,
		},
		{
			"download /dir/a !*",
			TestUser{"user", nil},
			false,
		},
		{
			"download /dir/a -user !*",
			TestUser{"user", nil},
			true,
		},
		{
			"download /dir/a =group !*",
			TestUser{"user", []string{"group"}},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.input,
			func(t *testing.T) {
				r, err := NewRule(tt.input)
				if err != nil {
					t.Errorf("unable to parse rule '%s': %s", tt.input, err)
					return
				}

				p, err := NewPermissions([]Rule{r})
				if err != nil {
					t.Errorf("unable to create Permissions: %s", err)
					return
				}

				allowed := p.Allowed(r.scope, r.path, tt.user)
				if allowed != tt.expected {
					t.Errorf("expected %t got %t", tt.expected, allowed)
					return
				}
			},
		)
	}
}
