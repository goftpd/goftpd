package acl

import "time"

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

type Group struct {
	Name string

	AddedAt time.Time
}

type GroupSettings struct {
	IsAdmin bool
	AddedAt time.Time
}
