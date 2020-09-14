package vfs

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/goftpd/goftpd/acl"
	"github.com/pkg/errors"
)

func TestNewFilesystemMakeDir(t *testing.T) {
	var tests = []struct {
		line string
		path string
		user TestUser
		err  error
	}{
		{
			"makedir / *",
			"/hello",
			newTestUser("user", "group"),
			nil,
		},
		{
			"makedir / !*",
			"/hello",
			newTestUser("user", "group"),
			acl.ErrPermissionDenied,
		},
		{
			"makedir / *",
			"/hello/something",
			newTestUser("user", "group"),
			errors.New("file does not exist"),
		},
		{
			"makedir / *",
			"/file/something",
			newTestUser("user", "group"),
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

				err := fs.MakeDir(tt.path, tt.user)
				checkErr(t, err, tt.err)

				if tt.err == nil {
					username, group, err := fs.shadow.Get(tt.path)
					if err != nil {
						t.Fatalf("expected nil but got '%s' for shadow.Get", err)
					}

					if username != tt.user.Name() {
						t.Errorf("expected shadow to be '%s' but got '%s'", tt.user.Name(), username)
					}

					if group != tt.user.PrimaryGroup() {
						t.Errorf("expected shadow to be '%s' but got '%s'", tt.user.PrimaryGroup(), group)
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
		user TestUser
		err  error
	}{
		{
			"/file",
			newTestUser("user"),
			nil,
		},
		{
			"/file2",
			newTestUser("user"),
			errors.New("file does not exist"),
		},
		{
			"/file",
			newTestUser("badUser"),
			errors.New("acl permission denied"),
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

				reader, err := fs.DownloadFile(tt.path, tt.user)
				checkErr(t, err, tt.err)

				if tt.err == nil {
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
		user    TestUser
		err     error
	}{
		{
			"/file",
			false,
			"HELLO",
			newTestUser("user"),
			nil,
		},
		{
			"/file",
			true,
			"HELLO",
			newTestUser("user"),
			errors.New("file already exists"),
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

				if tt.dupe {
					createFile(t, fs, tt.path, tt.content)
				}

				writer, err := fs.UploadFile(tt.path, tt.user)
				checkErr(t, err, tt.err)

				if tt.err == nil {

					fmt.Fprint(writer, tt.content)

					if err := writer.Close(); err != nil {
						t.Fatalf("unexpected err in close: %s", err)
					}

					username, group, err := fs.shadow.Get(tt.path)
					if err != nil {
						t.Fatalf("unexpected err in shadow.Get: %s", err)
					}

					if username != tt.user.Name() {
						t.Errorf("expected username to be '%s' got: '%s'", tt.user.Name(), username)
					}

					if group != "nobody" {
						t.Errorf("expected group to be nobody got: '%s'", group)
					}
				}
			},
		)
	}
}

func TestResumeUploadFile(t *testing.T) {
	var tests = []struct {
		create  bool
		path    string
		rules   []string
		content string
		owner   TestUser
		user    TestUser
		err     error
	}{
		{
			true,
			"/file",
			[]string{
				"resume / *",
				"upload / *",
			},
			"HELLO",
			newTestUser("user"),
			newTestUser("user"),
			nil,
		},
		{
			true,
			"/file",
			[]string{
				"resume / *",
			},
			"HELLO",
			newTestUser("user"),
			newTestUser("user"),
			acl.ErrPermissionDenied,
		},
		{
			true,
			"/file",
			[]string{
				"upload / *",
			},
			"HELLO",
			newTestUser("user"),
			newTestUser("user"),
			acl.ErrPermissionDenied,
		},
		{
			true,
			"/file",
			[]string{
				"resume / !*",
				"resumeown / *",
				"upload / *",
			},
			"HELLO",
			newTestUser("owner"),
			newTestUser("user"),
			acl.ErrPermissionDenied,
		},
		{
			true,
			"/file",
			[]string{
				"resume / !*",
				"resumeown / *",
				"upload / *",
			},
			"HELLO",
			newTestUser("owner"),
			newTestUser("owner"),
			nil,
		},
		{
			false,
			"/file",
			[]string{
				"resume / *",
				"upload / *",
			},
			"NOTHING TO RESUME",
			newTestUser("owner"),
			newTestUser("owner"),
			errors.New("file does not exist"),
		},
	}

	for idx, tt := range tests {
		t.Run(
			fmt.Sprintf("%d", idx),
			func(t *testing.T) {
				fs := newMemoryFilesystem(t, tt.rules)
				if fs == nil {
					t.Fatal("unexpected nil for fs")
				}
				defer stopMemoryFilesystem(t, fs)

				// create base file to resume
				if tt.create {
					createFile(t, fs, tt.path, tt.content)
					setShadowOwner(t, fs, tt.path, tt.owner)
				}

				writer, err := fs.ResumeUploadFile(tt.path, tt.user)
				checkErr(t, err, tt.err)

				if tt.err == nil {

					fmt.Fprint(writer, tt.content)

					if err := writer.Close(); err != nil {
						t.Fatalf("unexpected err in close: %s", err)
					}

					username, group, err := fs.shadow.Get(tt.path)
					if err != nil {
						t.Fatalf("unexpected err in shadow.Get: %s", err)
					}

					if username != tt.user.Name() {
						t.Errorf("expected username to be '%s' got: '%s'", tt.user.Name(), username)
					}

					if group != "nobody" {
						t.Errorf("expected group to be nobody got: '%s'", group)
					}

					// get file and make sure it is content * 2
					reader, err := fs.chroot.Open(tt.path)
					if err != nil {
						t.Fatalf("unexpected err in chroot.Open: %s", err)
					}

					b, err := ioutil.ReadAll(reader)
					if err != nil {
						t.Fatalf("expected nil reading file got: %s", err)
					}

					if string(b) != fmt.Sprintf("%s%s", tt.content, tt.content) {
						t.Fatalf("expected file to be '%s%s' but got: '%s'", tt.content, tt.content, string(b))
					}
				}
			},
		)
	}
}

func TestRenameFile(t *testing.T) {
	var tests = []struct {
		create  bool
		path    string
		newpath string
		rules   []string
		owner   TestUser
		user    TestUser
		err     error
	}{
		{
			true,
			"/file",
			"/file2",
			[]string{
				"rename / *",
				"upload / *",
			},
			newTestUser("user"),
			newTestUser("user"),
			nil,
		},
		{
			true,
			"/file",
			"/file2",
			[]string{
				"rename / !*",
				"upload / *",
			},
			newTestUser("user"),
			newTestUser("user"),
			acl.ErrPermissionDenied,
		},
		{
			true,
			"/file",
			"/file2",
			[]string{
				"rename / *",
				"upload / *",
			},
			newTestUser("owner"),
			newTestUser("user"),
			nil,
		},
		{
			true,
			"/file",
			"/file2",
			[]string{
				"rename / !*",
				"renameown / *",
				"upload / *",
			},
			newTestUser("user"),
			newTestUser("user"),
			nil,
		},
		{
			false,
			"/file",
			"/file2",
			[]string{
				"rename / *",
				"upload / *",
			},
			newTestUser("user"),
			newTestUser("user"),
			errors.New("file does not exist"),
		},
		{
			true,
			"/file",
			"/file",
			[]string{
				"rename / *",
				"upload / *",
			},
			newTestUser("user"),
			newTestUser("user"),
			errors.New("can not rename to self"),
		},
	}

	for idx, tt := range tests {
		t.Run(
			fmt.Sprintf("%d", idx),
			func(t *testing.T) {
				fs := newMemoryFilesystem(t, tt.rules)
				if fs == nil {
					t.Fatal("unexpected nil for fs")
				}
				defer stopMemoryFilesystem(t, fs)

				// create base file to resume
				if tt.create {
					createFile(t, fs, tt.path, "RENAME FILE")
					setShadowOwner(t, fs, tt.path, tt.owner)
				}

				err := fs.RenameFile(tt.path, tt.newpath, tt.user)
				checkErr(t, err, tt.err)

				if tt.err == nil {
					// check that shadow doesnt have the old path
					if _, _, err := fs.shadow.Get(tt.path); err != ErrNoPath {
						t.Fatalf("expected Get to be ErrNoPath got: %s", err)
					}

					// check that shadow has the new path
					username, group, err := fs.shadow.Get(tt.newpath)
					if err != nil {
						t.Fatalf("unexpected err in shadow.Get: %s", err)
					}

					if username != tt.user.Name() {
						t.Errorf("expected username to be '%s' got: '%s'", tt.user.Name(), username)
					}

					if group != "nobody" {
						t.Errorf("expected group to be nobody got: '%s'", group)
					}
				}
			},
		)
	}
}

func TestDeleteFile(t *testing.T) {
	var tests = []struct {
		create bool
		path   string
		rules  []string
		owner  TestUser
		user   TestUser
		err    error
	}{
		{
			true,
			"/file",
			[]string{
				"delete / *",
			},
			newTestUser("user"),
			newTestUser("user"),
			nil,
		},
		{
			false,
			"/file",
			[]string{
				"delete / *",
			},
			newTestUser("user"),
			newTestUser("user"),
			errors.New("file does not exist"),
		},
		{
			true,
			"/file",
			[]string{
				"delete / !*",
			},
			newTestUser("user"),
			newTestUser("user"),
			acl.ErrPermissionDenied,
		},
		{
			true,
			"/file",
			[]string{
				"delete / !*",
				"deleteown / *",
			},
			newTestUser("owner"),
			newTestUser("user"),
			acl.ErrPermissionDenied,
		},
		{
			true,
			"/file",
			[]string{
				"delete / !*",
				"deleteown / *",
			},
			newTestUser("owner"),
			newTestUser("owner"),
			nil,
		},
	}

	for idx, tt := range tests {
		t.Run(
			fmt.Sprintf("%d", idx),
			func(t *testing.T) {
				fs := newMemoryFilesystem(t, tt.rules)
				if fs == nil {
					t.Fatal("unexpected nil for fs")
				}
				defer stopMemoryFilesystem(t, fs)

				// create base file to resume
				if tt.create {
					createFile(t, fs, tt.path, "DELETE FILE")
					setShadowOwner(t, fs, tt.path, tt.owner)
				}

				err := fs.DeleteFile(tt.path, tt.user)
				checkErr(t, err, tt.err)

				if tt.err == nil {
					// check that shadow doesnt have the old path
					if _, _, err := fs.shadow.Get(tt.path); err != ErrNoPath {
						t.Fatalf("expected Get to be ErrNoPath got: %s", err)
					}
				}
			},
		)
	}
}
