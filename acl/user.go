package acl

type User struct {
	name   string
	groups []string
}

func (u User) Name() string { return u.name }
func (u User) PrimaryGroup() string {
	if len(u.groups) == 0 {
		return "nobody"
	}
	return u.groups[0]
}
func (u User) Groups() []string { return u.groups }

func NewUser(name string, groups []string) User {
	return User{
		name:   name,
		groups: groups,
	}
}
