// Package acl provides primitives for creating and checking permissions
// based on user and groups
package acl

import (
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

var AllowedUserAndGroupCharsRE = regexp.MustCompile(`[a-zA-Z0-9]`)

var ErrPermissionDenied = errors.New("acl permission denied")
var ErrBadInput = errors.New("bad input")

// collection is a container for the three different permission types,
// users and group. Provides utilities for checking if the collection
// contains a provided entity
type collection struct {
	all bool

	users  []string
	groups []string
}

// ACL provides utilities for checking if a subject has permission to perform
// on an object
type ACL struct {
	allowed collection
	blocked collection
}

// Takes in a string that describes the permissions for an object. Returns an ACL with
// a method for checking permissions. An entity is a user with the following attributes:
// - name
// - list of groups
//
// When describing permissions use the following (glftpd) syntax:
// - `-` prefix describes a user, i.e. `-userName`
// - `=` prefix describes a group, i.e. `=groupName`
// - no prefix describes a flag, i.e. `1` (currently no restrictions on legnth)
// - `!` prefix denotes that the preceding permission is blocked, i.e. `!-userName` would
// not be allowed
//
// Currently the order of checking is:
// - blocked users
// - blocked groups
// - allowed users
// - allowed groups
// - blocked all (!*)
// - allowed all (*)
//
// The default is to block permission
func NewFromString(s string) (*ACL, error) {
	if len(s) == 0 {
		return nil, ErrBadInput
	}

	var a ACL

	fields := strings.Fields(strings.ToLower(s))

	var c *collection

	for _, f := range fields {
		c = &a.allowed

		if f[0] == '!' {
			if len(f) <= 1 {
				return nil, errors.New("expected string after '!'")
			}

			c = &a.blocked

			f = f[1:]
		}

		switch f[0] {
		case '-':
			// user specific acl
			if len(f) <= 1 {
				return nil, errors.New("expected string after '-'")
			}

			f = f[1:]

			if f == "*" {
				return nil, errors.New("bad user '*'")
			}

			if !AllowedUserAndGroupCharsRE.MatchString(f) {
				return nil, errors.Errorf("user contains invalid characters: '%s'", f)
			}

			c.users = append(c.users, f)

		case '=':
			// group specific acl
			if len(f) <= 1 {
				return nil, errors.New("expected string after '='")
			}

			f = f[1:]

			if f == "*" {
				return nil, errors.New("bad group '*'")
			}

			if !AllowedUserAndGroupCharsRE.MatchString(f) {
				return nil, errors.Errorf("group contains invalid characters: '%s'", f)
			}

			c.groups = append(c.groups, f)

		default:
			if f != "*" {
				return nil, errors.Errorf("unexpected string in acl input: '%s'", f)
			}

			c.all = true
		}

	}

	return &a, nil
}

// has checks to see if the slice contains the provided element (lower cased)
func (c *collection) has(s []string, e string) bool {
	e = strings.ToLower(e)
	for idx := range s {
		if s[idx] == e {
			return true
		}
	}
	return false
}

// hasUser checks to see if the users slices contains the fgiven user
func (c *collection) hasUser(u string) bool {
	return c.has(c.users, u)
}

// hasGroup checks to see if the groups slice contains given group
func (c *collection) hasGroup(g string) bool {
	return c.has(c.groups, g)
}

// UserAllowed checks to see if given User is allowed or blocked. Default is to
// block access
func (a *ACL) Allowed(u *User) bool {
	// check blocked lists
	if a.blocked.hasUser(u.Name()) {
		return false
	}

	groups := u.Groups()
	for idx := range groups {
		if a.blocked.hasGroup(groups[idx]) {
			return false
		}
	}

	// check allowed lists
	if a.allowed.hasUser(u.Name()) {
		return true
	}

	for idx := range groups {
		if a.allowed.hasGroup(groups[idx]) {
			return true
		}
	}

	// fall back to catchalls '*' '!*'
	if a.blocked.all {
		return false
	}

	return a.allowed.all
}
