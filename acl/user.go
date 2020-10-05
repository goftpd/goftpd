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
	Groups       map[string]*GroupSettings

	// bytes available for download
	Ratio   int
	Credits int

	// login based attributes
	Logins    int
	Uploads   int
	Downloads int

	// meta
	AddedBy     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	LastLoginAt time.Time
	DeletedAt   time.Time

	IPMasks []string
}

func (u *User) HasGroup(name string) bool {
	name = strings.ToLower(name)

	_, ok := u.Groups[name]

	return ok
}

// AddGroup creates a new GroupSettings (rather than importing for every lua script)
func (u *User) AddGroup(name string) {
	name = strings.ToLower(name)

	if _, ok := u.Groups[name]; ok {
		return
	}

	u.Groups[name] = &GroupSettings{
		AddedAt: time.Now(),
	}
}

// AddGroup creates a new GroupSettings (rather than importing for every lua script)
func (u *User) RemoveGroup(name string) {
	name = strings.ToLower(name)

	if _, ok := u.Groups[name]; !ok {
		return
	}

	delete(u.Groups, name)
}

// Delete sets DeletedAt (convenience for scripts)
func (u *User) Delete() {
	u.DeletedAt = time.Now()
}

// Readd sets DeletedAt to nil (convenience for scripts)
func (u *User) Readd() {
	u.DeletedAt = time.Time{}
}

// AddIP attempts to validate and add an IP mask to a user
func (u *User) AddIP(mask string) error {
	mask = strings.ToLower(mask)

	parts := strings.Split(mask, "@")
	if len(parts) != 2 {
		return errors.New("does not contain a '@'")
	}

	if len(strings.Split(parts[1], ".")) != 4 {
		return errors.New("require 4 octets, '*' does not work across octets.")
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

// DeleteIP deletes the IP mask (if it exists) from the user
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

func (u *User) SetUpdatedAt() { u.UpdatedAt = time.Now() }

type GroupSettings struct {
	IsAdmin bool
	AddedAt time.Time
}
