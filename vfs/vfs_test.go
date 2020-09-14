package vfs

import (
	"fmt"
	"testing"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/goftpd/goftpd/acl"
	"github.com/pkg/errors"
)

type TestUser struct {
	name   string
	groups []string
}

func (u TestUser) Name() string {
	return u.name
}

func (u TestUser) Groups() []string {
	return u.groups
}

func (u TestUser) PrimaryGroup() string {
	if len(u.groups) > 0 {
		return u.groups[0]
	}
	return "nobody"
}

func newMemoryFilesystem(t *testing.T, lines []string) *Filesystem {
	memory := memfs.New()

	if err := memory.MkdirAll("/", defaultPerms); err != nil {
		t.Errorf("unexpected error creating root path: %s", err)
		return nil
	}

	ss := newMemoryShadowStore(t)

	var rules []acl.Rule
	for _, l := range lines {
		r, err := acl.NewRule(l)
		if err != nil {
			t.Errorf("unexpected error creating NewRules: %s", err)
			return nil
		}
		rules = append(rules, r)
	}

	perms, err := acl.NewPermissions(rules)
	if err != nil {
		t.Errorf("unexpected error creating Permissions: %s", err)
		return nil
	}

	fs, err := NewFilesystem(memory, ss, perms)
	if err != nil {
		t.Errorf("unexpected error creating NewFilesystem: %s", err)
		return nil
	}

	return fs
}

func stopMemoryFilesystem(t *testing.T, fs *Filesystem) {
	err := fs.Stop()
	if err != nil {
		t.Errorf("unexpected error stopping filesystem: %s", err)
		return
	}
}

func TestNewFilesystemMakeDir(t *testing.T) {
	var tests = []struct {
		line  string
		path  string
		user  string
		group string
		err   error
	}{
		{
			"makedir / *",
			"/hello",
			"user",
			"group",
			nil,
		},
		{
			"makedir / !*",
			"/hello",
			"user",
			"group",
			acl.ErrPermissionDenied,
		},
		{
			"makedir / *",
			"/hello/something",
			"user",
			"group",
			errors.New("file does not exist"),
		},
		{
			"makedir / *",
			"/file/something",
			"user",
			"group",
			errors.New("parent is not a directory"),
		},
	}

	for idx, tt := range tests {
		t.Run(
			fmt.Sprintf("%d", idx),
			func(t *testing.T) {
				fs := newMemoryFilesystem(t, []string{tt.line})
				if fs == nil {
					t.Error("unexpected nil for fs")
					return
				}
				defer stopMemoryFilesystem(t, fs)

				f, err := fs.chroot.Create("/file")
				if err != nil {
					t.Errorf("unexpected err creating /file: %s", err)
					return
				}

				fmt.Fprint(f, "HELLO")

				if err := f.Close(); err != nil {
					t.Errorf("unexpected err closing /file: %s", err)
					return
				}

				user := TestUser{tt.user, []string{tt.group}}

				err = fs.MakeDir(tt.path, user)
				if err == nil && tt.err != nil {
					t.Errorf("unexpected nil wanted: %s", tt.err)
					return
				}

				if err != nil && tt.err == nil {
					t.Errorf("expected nil but got: %s", err)
					return
				}

				if err != nil && tt.err != nil {
					if err.Error() != tt.err.Error() {
						t.Errorf("expected '%s' but got '%s'", tt.err, err)
						return
					}
				}

				if tt.err == nil {
					username, group, err := fs.shadow.Get(tt.path)
					if err != nil {
						t.Errorf("expected nil but got '%s' for shadow.Get", err)
						return
					}

					if username != tt.user {
						t.Errorf("expected shadow to be '%s' but got '%s'", tt.user, username)
						return
					}

					if group != tt.group {
						t.Errorf("expected shadow to be '%s' but got '%s'", tt.group, group)
						return
					}
				}
			},
		)
	}
}
