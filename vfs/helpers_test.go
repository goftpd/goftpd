package vfs

import (
	"fmt"
	"os"
	"testing"

	"github.com/dgraph-io/badger/v2"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/goftpd/goftpd/acl"
)

func newMemoryShadowStore(t *testing.T) Shadow {
	opt := badger.DefaultOptions("").WithInMemory(true)

	db, err := badger.Open(opt)
	if err != nil {
		t.Fatalf("error opening db: %s", err)
	}

	ss := NewShadowStore(db)

	return ss
}

func closeMemoryShadowStore(t *testing.T, ss Shadow) {
	if err := ss.Close(); err != nil {
		t.Fatalf("error closing shadow: %s", err)
	}
}

func newTestUser(name string, groups ...string) *acl.User {
	u := acl.NewUser(name, groups)
	return &u
}

func checkErr(t *testing.T, got, expected error) {
	if got == nil {
		if expected != nil {
			t.Fatalf("expected '%s' but got nil", expected)
			return
		}
		return
	}

	if expected == nil {
		t.Fatalf("unexpected error '%s'", got)
		return
	}
}

func createFile(t *testing.T, fs *Filesystem, path, contents string) {
	f, err := fs.chroot.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
	if err != nil {
		t.Fatalf("unexpected err creating %s: %s", path, err)
	}

	fmt.Fprint(f, contents)

	if err := f.Close(); err != nil {
		t.Fatalf("unexpected err closing %s: %s", path, err)
	}
}

func setShadowOwner(t *testing.T, fs *Filesystem, path string, owner *acl.User) {
	if err := fs.shadow.Set(path, owner.Name(), owner.PrimaryGroup()); err != nil {
		t.Fatalf("unexpected err setting shadow owner: %s", err)
	}
}

func newMemoryFilesystem(t *testing.T, lines []string) *Filesystem {
	memory := memfs.New()

	if err := memory.MkdirAll("/", defaultPerms); err != nil {
		t.Fatalf("unexpected error creating root path: %s", err)
	}

	ss := newMemoryShadowStore(t)

	var rules []acl.Rule
	for _, l := range lines {
		r, err := acl.NewRule(l)
		if err != nil {
			t.Fatalf("unexpected error creating NewRules: %s", err)
		}
		rules = append(rules, r)
	}

	perms, err := acl.NewPermissions(rules)
	if err != nil {
		t.Fatalf("unexpected error creating Permissions: %s", err)
	}

	fs, err := NewFilesystem(memory, ss, perms)
	if err != nil {
		t.Fatalf("unexpected error creating NewFilesystem: %s", err)
	}

	return fs
}

func stopMemoryFilesystem(t *testing.T, fs *Filesystem) {
	err := fs.Stop()
	if err != nil {
		t.Fatalf("unexpected error stopping filesystem: %s", err)
	}
}
