package vfs

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/goftpd/goftpd/acl"
	"github.com/pkg/errors"
)

var defaultPerms os.FileMode = 0666

type Filesystem struct {
	chroot      billy.Filesystem
	shadow      Shadow
	permissions *acl.Permissions
}

// NewFilesystem creates a new Filesystem with the given chroot (underlying fs) shadow (stores user/group meta data
// and permissions (check acl for paths, users and different scopes)
func NewFilesystem(chroot billy.Filesystem, shadow Shadow, permissions *acl.Permissions) (*Filesystem, error) {
	fs := Filesystem{
		chroot:      chroot,
		shadow:      shadow,
		permissions: permissions,
	}

	return &fs, nil
}

func (fs *Filesystem) Stop() error {
	if err := fs.shadow.Close(); err != nil {
		return err
	}
	return nil
}

// MakeDir checks to see if the user has permission to create a new directory. Does so if allowed
func (fs *Filesystem) MakeDir(path string, user acl.User) error {
	if !fs.permissions.Allowed(acl.PermissionScopeMakeDir, path, user) {
		return acl.ErrPermissionDenied
	}

	// make sure the base exists and is a directory
	path = filepath.Clean(path)
	dir := filepath.Dir(path)

	finfo, err := fs.chroot.Stat(dir)
	if err != nil {
		return err
	}

	if !finfo.IsDir() {
		return errors.New("parent is not a directory")
	}

	if err := fs.chroot.MkdirAll(path, defaultPerms); err != nil {
		return err
	}

	if err := fs.shadow.Set(path, user.Name(), user.PrimaryGroup()); err != nil {
		return err
	}

	return nil
}

// DownloadFile checks to see if the user has permission to read the file (checking download
// permissions from high level to low level). Returns an io.Reader if allowed
func (fs *Filesystem) DownloadFile(path string, user acl.User) (io.Reader, error) {
	if !fs.permissions.Allowed(acl.PermissionScopeDownload, path, user) {
		return nil, acl.ErrPermissionDenied
	}

	f, err := fs.chroot.Open(path)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// UploadFile checks to see if the user has permission to write the file (checking upload
// permissions from high level to low level). Returns an io.Writer if allowed. Does not
// truncate a file
func (fs *Filesystem) UploadFile(path string, user acl.User) (io.WriteCloser, error) {
	if !fs.permissions.Allowed(acl.PermissionScopeUpload, path, user) {
		return nil, acl.ErrPermissionDenied
	}

	f, err := fs.chroot.OpenFile(path, os.O_RDWR|os.O_CREATE, defaultPerms)
	if err != nil {
		return nil, err
	}

	// wrap the file in our special Writer that allows us to manage the shadow fs
	writer := newWriteCloser(f, func() error {
		return fs.shadow.Set(path, user.Name(), user.PrimaryGroup())
	})

	return writer, nil
}

// ResumeUploadFile checks to see if the user has permission to write the file (checking upload
// permissions from high level to low level). It also checks to see if they have resume writes.
// Returns an io.Writer if allowed.
func (fs *Filesystem) ResumeUploadFile(path string, user acl.User) (io.WriteCloser, error) {
	if !fs.permissions.Allowed(acl.PermissionScopeUpload, path, user) {
		return nil, acl.ErrPermissionDenied
	}

	if !fs.permissions.Allowed(acl.PermissionScopeResume, path, user) {
		// not allowed to globally resume, check if this is ours and we can resume our own
		if !fs.permissions.Allowed(acl.PermissionScopeResumeOwn, path, user) {
			return nil, acl.ErrPermissionDenied
		}

		owner, err := fs.checkOwnership(path, user)
		if err != nil {
			return nil, err
		}

		if !owner {
			return nil, acl.ErrPermissionDenied
		}
	}

	f, err := fs.chroot.OpenFile(path, os.O_RDWR|os.O_APPEND, defaultPerms)
	if err != nil {
		return nil, err
	}

	if _, err := f.Seek(0, os.SEEK_END); err != nil {
		return nil, err
	}

	// wrap the file in our special Writer that allows us to manage the shadow fs
	writer := newWriteCloser(f, func() error {
		return fs.shadow.Set(path, user.Name(), user.PrimaryGroup())
	})

	return writer, nil
}

// RenameFile checks to see if the user has permission to rename the file (checking rename and
// renameown scopes).
func (fs *Filesystem) RenameFile(oldpath, newpath string, user acl.User) error {
	// need a way to transform usernames to uid and groups to gid, shadowing the entire
	// fs is not ideal

	// make sure that the user has permission to upload to the new path
	if !fs.permissions.Allowed(acl.PermissionScopeUpload, newpath, user) {
		return acl.ErrPermissionDenied
	}

	if !fs.permissions.Allowed(acl.PermissionScopeRename, oldpath, user) {

		// not allowed to globally rename, check if this is ours and we can rename our own
		if !fs.permissions.Allowed(acl.PermissionScopeRenameOwn, oldpath, user) {
			return acl.ErrPermissionDenied
		}

		owner, err := fs.checkOwnership(oldpath, user)
		if err != nil {
			return err
		}

		if !owner {
			return acl.ErrPermissionDenied
		}
	}

	if err := fs.chroot.Rename(oldpath, newpath); err != nil {
		return err
	}

	if err := fs.shadow.Remove(oldpath); err != nil {
		return err
	}

	if err := fs.shadow.Set(newpath, user.Name(), user.PrimaryGroup()); err != nil {
		return err
	}

	return nil
}

// DeleteFile checks to see if the user has permission to delete the file (checking delete and
// deleteown scopes).
func (fs *Filesystem) DeleteFile(path string, user acl.User) error {
	// need a way to transform usernames to uid and groups to gid, shadowing the entire
	// fs is not ideal

	if !fs.permissions.Allowed(acl.PermissionScopeDelete, path, user) {

		// not allowed to globally delete, check if this is ours and we can delete our own
		if !fs.permissions.Allowed(acl.PermissionScopeDeleteOwn, path, user) {
			return acl.ErrPermissionDenied
		}

		owner, err := fs.checkOwnership(path, user)
		if err != nil {
			return err
		}

		if !owner {
			return acl.ErrPermissionDenied
		}
	}

	if err := fs.chroot.Remove(path); err != nil {
		return err
	}

	if err := fs.shadow.Remove(path); err != nil {
		return err
	}

	return nil
}

// checkOwnership checks to see if a user is an owner of a given path. Returns bool
// and an error
func (fs *Filesystem) checkOwnership(path string, user acl.User) (bool, error) {
	username, _, err := fs.shadow.Get(path)
	if err != nil {
		return false, err
	}

	if username != strings.ToLower(user.Name()) {
		return false, nil
	}

	return true, nil
}
