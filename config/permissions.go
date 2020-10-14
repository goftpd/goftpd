package config

import (
	"github.com/goftpd/goftpd/acl"
	"github.com/pkg/errors"
)

func (c *Config) ParsePermissions() (*acl.Permissions, error) {
	lines, ok := c.lines[NamespaceACL]
	if !ok {
		return nil, errors.New("no acl options provided")
	}

	var rules []acl.Rule
	for _, l := range lines {
		r, err := acl.NewRule(l.text)
		if err != nil {
			return nil, errors.Errorf("error parsing acl rule on line %d: %s", l.line, err)
		}
		rules = append(rules, r)
	}

	permissions := acl.NewPermissions(rules)

	return permissions, nil
}
