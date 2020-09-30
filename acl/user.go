package acl

import (
	"fmt"
	"strings"
	"time"
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

func (u *User) AddIP(mask string) bool {
	mask = strings.ToLower(mask)

	var match bool
	for idx := range u.IPMasks {
		if mask == u.IPMasks[idx] {
			match = true
			break
		}
	}
	if match {
		return false
	}

	u.IPMasks = append(u.IPMasks, mask)

	return true
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
