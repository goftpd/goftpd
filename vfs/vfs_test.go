package vfs

import (
	"fmt"
	"io/ioutil"
	"os"
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
					t.Fatal("unexpected nil for fs")
				}
				defer stopMemoryFilesystem(t, fs)

				createFile(t, fs, "/file", "HELLO")

				user := TestUser{tt.user, []string{tt.group}}

				err := fs.MakeDir(tt.path, user)
				if err == nil && tt.err != nil {
					t.Fatalf("unexpected nil wanted: %s", tt.err)
				}

				if err != nil && tt.err == nil {
					t.Fatalf("expected nil but got: %s", err)
				}

				if err != nil && tt.err != nil {
					if err.Error() != tt.err.Error() {
						t.Fatalf("expected '%s' but got '%s'", tt.err, err)
					}
				}

				if tt.err == nil {
					username, group, err := fs.shadow.Get(tt.path)
					if err != nil {
						t.Fatalf("expected nil but got '%s' for shadow.Get", err)
					}

					if username != tt.user {
						t.Fatalf("expected shadow to be '%s' but got '%s'", tt.user, username)
					}

					if group != tt.group {
						t.Fatalf("expected shadow to be '%s' but got '%s'", tt.group, group)
					}
				}
			},
		)
	}
}

func TestDownloadFile(t *testing.T) {
	var rule = "download / !-badUser *"

	var tests = []struct {
		path string
		user string
		err  string
	}{
		{
			"/file",
			"user",
			"",
		},
		{
			"/file2",
			"user",
			"file does not exist",
		},
		{
			"/file",
			"badUser",
			"acl permission denied",
		},
	}

	for idx, tt := range tests {
		t.Run(
			fmt.Sprintf("%d", idx),
			func(t *testing.T) {
				fs := newMemoryFilesystem(t, []string{rule})
				if fs == nil {
					t.Fatal("unexpected nil for fs")
				}
				defer stopMemoryFilesystem(t, fs)

				createFile(t, fs, "/file", "HELLO")

				var testUser = TestUser{tt.user, nil}

				reader, err := fs.DownloadFile(tt.path, testUser)
				if err != nil && len(tt.err) == 0 {
					t.Fatalf("expected nil got '%s'", err)
				}

				if err != nil && err.Error() != tt.err {
					t.Fatalf("expected '%s' got '%s'", tt.err, err)
				}

				if err == nil && len(tt.err) > 0 {
					t.Fatalf("expected '%s' got nil", tt.err)
				}

				if len(tt.err) == 0 {
					defer reader.Close()

					b, err := ioutil.ReadAll(reader)
					if err != nil {
						t.Fatalf("expected nil reading file got: %s", err)
					}

					if string(b) != "HELLO" {
						t.Fatalf("got '%s' when we expected 'HELLO'", string(b))
					}
				}
			},
		)
	}
}

func TestUploadFile(t *testing.T) {
	var rule = "upload / !-badUser *"

	var tests = []struct {
		path    string
		dupe    bool
		content string
		user    string
		err     string
	}{
		{
			"/file",
			false,
			"HELLO",
			"user",
			"",
		},
		{
			"/file",
			true,
			"HELLO",
			"user",
			"file already exists",
		},
	}

	t.Log("begin testUploadFile")

	for idx, tt := range tests {
		t.Run(
			fmt.Sprintf("%d", idx),
			func(t *testing.T) {
				fs := newMemoryFilesystem(t, []string{rule})
				if fs == nil {
					t.Fatal("unexpected nil for fs")
				}
				defer stopMemoryFilesystem(t, fs)

				if tt.dupe {
					createFile(t, fs, tt.path, tt.content)
				}

				var testUser = TestUser{tt.user, nil}

				writer, err := fs.UploadFile(tt.path, testUser)

				t.Logf("writer: %+v err: %+v", writer, err)

				if err != nil && len(tt.err) == 0 {
					t.Fatalf("expected nil got '%s'", err)
				}

				if err != nil && err.Error() != tt.err {
					t.Fatalf("expected '%s' got '%s'", tt.err, err)
				}

				if err == nil && len(tt.err) > 0 {
					t.Fatalf("expected '%s' got nil", tt.err)
				}

				if err == nil && len(tt.content) > 0 {

					fmt.Fprint(writer, tt.content)

					if err := writer.Close(); err != nil {
						t.Fatalf("unexpected err in close: %s", err)
					}

					username, group, err := fs.shadow.Get(tt.path)
					if err != nil {
						t.Fatalf("unexpected err in shadow.Get: %s", err)
					}

					if username != tt.user {
						t.Fatalf("expected username to be '%s' got: '%s'", tt.user, username)
					}

					if group != "nobody" {
						t.Fatalf("expected group to be nobody got: '%s'", group)
					}
				}
			},
		)
	}
}
