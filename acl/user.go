package acl

import (
	"fmt"
	"strings"
	"time"

	"github.com/gobwas/glob"
	"github.com/pkg/errors"
)

type User struct {
	Name     string
	Password []byte

	// group related attributes
	PrimaryGroup string
	Groups       map[string]GroupSettings

	// bytes available for download
	Credits int

	// login based attributes
	Logins    int
	Uploads   int
	Downloads int

	// meta
	CreatedAt   time.Time
	LastLoginAt time.Time
	DeletedAt   time.Time

	IPMasks []string
}

func (u *User) AddIP(mask string) error {
	mask = strings.ToLower(mask)

	parts := strings.Split(mask, "@")
	if len(parts) != 2 {
		return errors.New("does not contain a '@'")
	}

	// check we can compile the glob
	_, err := glob.Compile(parts[0], '.')
	if err != nil {
		return errors.New("mask is not a valid 'glob'")
	}

	// TODO
	// not sure how to validate an octet string
	// here

	// TODO
	// minimum security needs to be checked, will need
	// to make this a private function called via Auth which
	// passes through a config/the minimum security details

	for idx := range u.IPMasks {
		if mask == u.IPMasks[idx] {
			return errors.New("mask already exists")
		}
	}

	u.IPMasks = append(u.IPMasks, mask)

	return nil
}

func (u *User) DeleteIP(mask string) bool {
	mask = strings.ToLower(mask)

	original := len(u.IPMasks)

	var idx int
	for _, m := range u.IPMasks {
		if m != mask {
			u.IPMasks[idx] = m
			idx++
		}
	}

	u.IPMasks = u.IPMasks[:idx]

	return original != len(u.IPMasks)
}

// Used to satisfy the authenticator Entry interface
func (u User) Key() []byte {
	return []byte(fmt.Sprintf(
		"users:%s",
		strings.ToLower(u.Name),
	))
}

type Group struct {
	Name string

	AddedAt time.Time
}

func (g Group) Key() []byte {
	return []byte(fmt.Sprintf(
		"groups:%s",
		strings.ToLower(g.Name),
	))
}

type GroupSettings struct {
	IsAdmin bool
	AddedAt time.Time
}
