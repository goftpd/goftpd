package acl

import (
	"strings"

	"github.com/pkg/errors"
)

// Rule represents a permission parsed from a config file
type Rule struct {
	path  string
	scope PermissionScope
	acl   *ACL
}

// NewRule takes a line of text (i.e. from a config file) and performs some
// validation
func NewRule(line string) (Rule, error) {
	var rule Rule

	fields := strings.Fields(strings.ToLower(line))

	if len(fields) < 3 {
		return rule, errors.New("rule requires minimum of 3 fields")
	}

	scope, ok := StringToPermissionScope[fields[0]]
	if !ok {
		return rule, errors.Errorf("unknown permission scope '%s'", fields[0])
	}
	rule.scope = scope

	rule.path = fields[1]

	acl, err := NewFromString(strings.Join(fields[2:], " "))
	if err != nil {
		return rule, err
	}
	rule.acl = acl

	return rule, nil
}

// Permissions is a snapshot of the current permissions. They are stored
// as PermissionScope and then path
type Permissions struct {
	current map[PermissionScope]map[string]*ACL
}

// NewPermissions takes a slice of Rules and creates a way for callers to check ACL
// for a given path and scope
func NewPermissions(rules []Rule) (*Permissions, error) {
	p := Permissions{
		current: make(map[PermissionScope]map[string]*ACL, 0),
	}

	for _, r := range rules {
		s, ok := p.current[r.scope]
		if !ok {
			s = make(map[string]*ACL, 0)
			p.current[r.scope] = s
		}

		// if path exists already for this scope error out
		if _, ok := s[r.path]; ok {
			return nil, errors.Errorf("path '%s' for scope '%s' already exists", r.path, r.scope)
		}

		s[r.path] = r.acl
	}

	return &p, nil
}

// Allowed takes a scope a path and a *User and checks to see if they are allowed or blocked based on the
// underlying ACL. Returns bool if they are allowed. Defaults to not allowing.
func (p *Permissions) Allowed(scope PermissionScope, path string, user *User) bool {
	s, ok := p.current[scope]
	if !ok {
		// potential to return an error here
		return false
	}

	path = strings.ToLower(path)

	// all paths are represented as unix, although they can be run on windows
	parts := strings.Split(path, "/")

	for i := len(parts); i >= 0; i-- {
		if acl, ok := s[strings.Join(parts[:i], "/")]; ok {
			return acl.Allowed(user)
		}
	}

	// check "/" explicitly
	if acl, ok := s["/"]; ok {
		return acl.Allowed(user)
	}

	return false
}
