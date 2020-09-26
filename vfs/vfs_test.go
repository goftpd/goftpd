package vfs

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/goftpd/goftpd/acl"
	"github.com/pkg/errors"
)

func TestNewFilesystemMakeDir(t *testing.T) {
	var tests = []struct {
		line string
		path string
		user *acl.User
		err  error
	}{
		{
			"makedir /** *",
			"/hello",
			newTestUser("user", "group"),
			nil,
		},
		{
			"makedir /** !*",
			"/hello",
			newTestUser("user", "group"),
			acl.ErrPermissionDenied,
		},
		{
			"makedir /** *",
			"/hello/something",
			newTestUser("user", "group"),
			errors.New("file does not exist"),
		},
		{
			"makedir /** *",
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

					if username != tt.user.Name {
						t.Errorf("expected shadow to be '%s' but got '%s'", tt.user.Name, username)
					}

					if group != tt.user.PrimaryGroup {
						t.Errorf("expected shadow to be '%s' but got '%s'", tt.user.PrimaryGroup, group)
					}
				}
			},
		)
	}
}

func TestDownloadFile(t *testing.T) {
	var rule = "download /** !-badUser *"

	var tests = []struct {
		path string
		user *acl.User
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
	var rule = "upload /** !-badUser *"

	var tests = []struct {
		path    string
		dupe    bool
		content string
		user    *acl.User
		err     error
	}{
		{
			"/file",
			false,
			"HELLO",
			newTestUser("user", "nobody"),
			nil,
		},
		{
			"/file",
			true,
			"HELLO",
			newTestUser("user", "nobody"),
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

					if username != tt.user.Name {
						t.Errorf("expected username to be '%s' got: '%s'", tt.user.Name, username)
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
		owner   *acl.User
		user    *acl.User
		err     error
	}{
		{
			true,
			"/file",
			[]string{
				"resume /** *",
				"upload /** *",
			},
			"HELLO",
			newTestUser("user", "nobody"),
			newTestUser("user", "nobody"),
			nil,
		},
		{
			true,
			"/file",
			[]string{
				"resume /** *",
			},
			"HELLO",
			newTestUser("user", "nobody"),
			newTestUser("user", "nobody"),
			acl.ErrPermissionDenied,
		},
		{
			true,
			"/file",
			[]string{
				"upload /** *",
			},
			"HELLO",
			newTestUser("user", "nobody"),
			newTestUser("user", "nobody"),
			acl.ErrPermissionDenied,
		},
		{
			true,
			"/file",
			[]string{
				"resume /** !*",
				"resumeown /** *",
				"upload /** *",
			},
			"HELLO",
			newTestUser("owner", "nobody"),
			newTestUser("user", "nobody"),
			acl.ErrPermissionDenied,
		},
		{
			true,
			"/file",
			[]string{
				"resume /** !*",
				"resumeown /** *",
				"upload /** *",
			},
			"HELLO",
			newTestUser("owner", "nobody"),
			newTestUser("owner", "nobody"),
			nil,
		},
		{
			false,
			"/file",
			[]string{
				"resume /** *",
				"upload /** *",
			},
			"NOTHING TO RESUME",
			newTestUser("owner", "nobody"),
			newTestUser("owner", "nobody"),
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

					if username != tt.user.Name {
						t.Errorf("expected username to be '%s' got: '%s'", tt.user.Name, username)
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
		owner   *acl.User
		user    *acl.User
		err     error
	}{
		{
			true,
			"/file",
			"/file2",
			[]string{
				"rename /** *",
				"upload /** *",
			},
			newTestUser("user", "nobody"),
			newTestUser("user", "nobody"),
			nil,
		},
		{
			true,
			"/file",
			"/file2",
			[]string{
				"rename /** !*",
				"upload /** *",
			},
			newTestUser("user", "nobody"),
			newTestUser("user", "nobody"),
			acl.ErrPermissionDenied,
		},
		{
			true,
			"/file",
			"/file2",
			[]string{
				"rename /** *",
				"upload /** *",
			},
			newTestUser("owner", "nobody"),
			newTestUser("user", "nobody"),
			nil,
		},
		{
			true,
			"/file",
			"/file2",
			[]string{
				"rename /** !*",
				"renameown /** *",
				"upload /** *",
			},
			newTestUser("user", "nobody"),
			newTestUser("user", "nobody"),
			nil,
		},
		{
			false,
			"/file",
			"/file2",
			[]string{
				"rename /** *",
				"upload /** *",
			},
			newTestUser("user", "nobody"),
			newTestUser("user", "nobody"),
			errors.New("file does not exist"),
		},
		{
			true,
			"/file",
			"/file",
			[]string{
				"rename /** *",
				"upload /** *",
			},
			newTestUser("user", "nobody"),
			newTestUser("user", "nobody"),
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

					if username != tt.user.Name {
						t.Errorf("expected username to be '%s' got: '%s'", tt.user.Name, username)
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
		owner  *acl.User
		user   *acl.User
		err    error
	}{
		{
			true,
			"/file",
			[]string{
				"delete /** *",
			},
			newTestUser("user", "nobody"),
			newTestUser("user", "nobody"),
			nil,
		},
		{
			false,
			"/file",
			[]string{
				"delete /** *",
			},
			newTestUser("user", "nobody"),
			newTestUser("user", "nobody"),
			errors.New("file does not exist"),
		},
		{
			true,
			"/file",
			[]string{
				"delete /** !*",
			},
			newTestUser("user", "nobody"),
			newTestUser("user", "nobody"),
			acl.ErrPermissionDenied,
		},
		{
			true,
			"/file",
			[]string{
				"delete /** !*",
				"deleteown /** *",
			},
			newTestUser("owner", "nobody"),
			newTestUser("user", "nobody"),
			acl.ErrPermissionDenied,
		},
		{
			true,
			"/file",
			[]string{
				"delete /** !*",
				"deleteown /** *",
			},
			newTestUser("owner", "nobody"),
			newTestUser("owner", "nobody"),
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

func TestListDirSortByName(t *testing.T) {
	owner := newTestUser("user", "group")
	rules := []string{
		"download /** *",
	}

	// potential to create random fails, but lets see
	now := time.Now()
	expectedDetailed := fmt.Sprintf(
		"-rw-rw-rw- 1 user group            9 %s f0\r\n-rw-rw-rw- 1 user group            9 %s f1\r\n-rw-rw-rw- 1 user group            9 %s f2\r\n-rw-rw-rw- 1 user group            9 %s f3\r\n-rw-rw-rw- 1 user group            9 %s f4\r\n-rw-rw-rw- 1 user group            9 %s f5\r\n-rw-rw-rw- 1 user group            9 %s f6\r\n-rw-rw-rw- 1 user group            9 %s f7\r\n-rw-rw-rw- 1 user group            9 %s f8\r\n-rw-rw-rw- 1 user group            9 %s f9\r\n",
		now.Format("Jan _2 15:04"),
		now.Format("Jan _2 15:04"),
		now.Format("Jan _2 15:04"),
		now.Format("Jan _2 15:04"),
		now.Format("Jan _2 15:04"),
		now.Format("Jan _2 15:04"),
		now.Format("Jan _2 15:04"),
		now.Format("Jan _2 15:04"),
		now.Format("Jan _2 15:04"),
		now.Format("Jan _2 15:04"),
	)

	fs := newMemoryFilesystem(t, rules)
	if fs == nil {
		t.Fatal("unexpected nil for fs")
	}
	defer stopMemoryFilesystem(t, fs)

	for i := 0; i < 10; i++ {
		path := fmt.Sprintf("/f%d", i)
		createFile(t, fs, path, "LIST FILE")
		setShadowOwner(t, fs, path, owner)
	}

	files, err := fs.ListDir("/", owner)
	checkErr(t, err, nil)

	if len(files) != 10 {
		t.Errorf("expected 10 files found %d", len(files))
	}

	files.SortByName()

	if string(files.Detailed()) != expectedDetailed {
		t.Errorf("unexpected detailed:\n%s\nexpected:\n%s", string(files.Detailed()), expectedDetailed)
	}

}
func TestListDirNoPermission(t *testing.T) {
	owner := newTestUser("user", "group")
	rules := []string{
		"download /** !*",
	}

	fs := newMemoryFilesystem(t, rules)
	if fs == nil {
		t.Fatal("unexpected nil for fs")
	}
	defer stopMemoryFilesystem(t, fs)

	for i := 0; i < 10; i++ {
		path := fmt.Sprintf("/f%d", i)
		createFile(t, fs, path, "LIST FILE")
		setShadowOwner(t, fs, path, owner)
	}

	files, err := fs.ListDir("/", owner)
	checkErr(t, err, acl.ErrPermissionDenied)

	if files != nil {
		t.Fatal("expected files to be nil")
	}
}
