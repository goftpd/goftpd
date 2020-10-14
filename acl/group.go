package acl

import (
	"fmt"
	"strings"
	"time"
)

type Group struct {
	Name        string
	Description string

	Slots      int
	LeechSlots int

	Users map[string]*UserGroupMeta

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (g Group) Key() []byte {
	return []byte(fmt.Sprintf(
		"groups:%s",
		strings.ToLower(g.Name),
	))
}

func (g *Group) SetUpdatedAt() { g.UpdatedAt = time.Now() }

func (g *Group) AddUser(caller, target string) bool {
	if len(g.Users) >= g.Slots {
		return false
	}

	target = strings.ToLower(target)

	if _, ok := g.Users[target]; ok {
		// already added
		return true
	}

	g.Users[target] = &UserGroupMeta{
		AddedBy: caller,
		AddedAt: time.Now(),
	}

	return true
}

func (g *Group) RemoveUser(target string) bool {
	target = strings.ToLower(target)

	if _, ok := g.Users[target]; !ok {
		return false
	}

	delete(g.Users, target)

	return true
}

type UserGroupMeta struct {
	AddedBy string
	AddedAt time.Time
}
