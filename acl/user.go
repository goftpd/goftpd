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

	// map of ident@ip matches against the time they were added
	// potential to add TTL on ips here, or for maintenace (clean
	// all ips older than x)
	IPs map[string]time.Time
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
