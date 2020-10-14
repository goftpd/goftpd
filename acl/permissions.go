package acl

import (
	"sort"
	"strings"

	"github.com/gobwas/glob"
	"github.com/pkg/errors"
)

var (
	ErrRuleInvalidInput           = errors.New("rule requires minimum of 3 fields")
	ErrRuleUnknownPermissionScope = errors.New("unknown permission scope")
	ErrRuleBadGlob                = errors.New("failed to compile glob")
)

// Rule represents a permission parsed from a config file
type Rule struct {
	path  string
	scope PermissionScope
	g     glob.Glob
	acl   *ACL
}

// NewRule takes a line of text (i.e. from a config file) and performs some
// validation
func NewRule(line string) (Rule, error) {
	var rule Rule

	fields := strings.Fields(strings.ToLower(line))

	if len(fields) < 3 {
		return rule, ErrRuleInvalidInput
	}

	scope, ok := StringToPermissionScope[fields[0]]
	if !ok {
		return rule, ErrRuleUnknownPermissionScope
	}
	rule.scope = scope

	rule.path = fields[1]

	g, err := glob.Compile(rule.path, '/')
	if err != nil {
		return rule, ErrRuleBadGlob
	}

	rule.g = g

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
	current map[PermissionScope][]Rule
}

// NewPermissions takes a slice of Rules and creates a way for callers to check ACL
// for a given path and scope
func NewPermissions(rules []Rule) *Permissions {
	p := Permissions{
		current: make(map[PermissionScope][]Rule, 0),
	}

	for _, r := range rules {
		s, ok := p.current[r.scope]
		if !ok {
			s = make([]Rule, 0)
			p.current[r.scope] = s
		}

		p.current[r.scope] = append(p.current[r.scope], r)
	}

	for k := range p.current {
		sort.Slice(p.current[k], func(i, j int) bool {
			return len(p.current[k][j].path) < len(p.current[k][i].path)
		})
	}

	return &p
}

// Match takes a scope a path and a *User and checks to see if they match any rules defaults
// to no match
func (p *Permissions) Match(scope PermissionScope, path string, user *User) bool {
	s, ok := p.current[scope]
	if !ok {
		// potential to return an error here
		return false
	}

	path = strings.ToLower(path)

	for _, r := range s {

		if r.g.Match(path) {
			return r.acl.Match(user)
		}
	}

	return false
}

// MatchNoDefault takes a scope a path and a *User and checks to see if they match any rules
func (p *Permissions) MatchNoDefault(scope PermissionScope, path string, user *User) (bool, bool) {
	s, ok := p.current[scope]
	if !ok {
		// potential to return an error here
		return false, false
	}

	path = strings.ToLower(path)

	for _, r := range s {

		if r.g.Match(path) {
			result, match := r.acl.ExplicitMatch(user)
			if match {
				return result, true
			}
		}
	}

	return false, false
}
