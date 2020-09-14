package vfs

import (
	"errors"
	"io"
	"os"

	"github.com/go-git/go-billy/v5"
	"github.com/goftpd/goftpd/acl"
)

type Filesystem struct {
	chroot      billy.Filesystem
	shadow      Shadow
	permissions acl.Permissions
}

func New(chroot billy.Filesystem, permissions acl.Permissions) (*Filesystem, error) {
	fs := Filesystem{
		chroot:      chroot,
		permissions: permissions,
	}

	return &fs, nil
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
func (fs *Filesystem) UploadFile(path string, user acl.User) (io.Writer, error) {
	if !fs.permissions.Allowed(acl.PermissionScopeUpload, path, user) {
		return nil, acl.ErrPermissionDenied
	}

	f, err := fs.chroot.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return f, nil
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

	return nil
}

// checkOwnership checks to see if a user is an owner of a given path. Returns bool
// and an error
func (fs *Filesystem) checkOwnership(path string, user acl.User) (bool, error) {
	// stat the file to check ownership
	// finfo, err := fs.chroot.Stat(path)
	// if err != nil {
	// 	return false, err
	// }

	// if finfo.User != user {
	// 		return nil, acl.ErrPermissionDenied
	// }

	return false, errors.New("stub")
}
