package vfs

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/go-git/go-billy/v5"
	"github.com/goftpd/goftpd/acl"
	"github.com/pkg/errors"
)

var defaultPerms os.FileMode = 0666

type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

func newBufferPoolWithSize(size int) sync.Pool {
	return sync.Pool{
		New: func() interface{} { s := make([]byte, size); return &s },
	}
}

type VFS interface {
	Join(string, []string) string
	Stop() error
	MakeDir(string, *acl.User) error
	DownloadFile(string, *acl.User) (ReadSeekCloser, int64, error)
	UploadFile(string, *acl.User) (io.WriteCloser, error)
	ResumeUploadFile(string, *acl.User) (io.WriteCloser, error)
	RenameFile(string, string, *acl.User) error
	DeleteFile(string, *acl.User) error
	DeleteDir(string, *acl.User) error
	ListDir(string, *acl.User) (FileList, error)

	GetBuffer() *[]byte
	PutBuffer(*[]byte)
}

type FilesystemOpts struct {
	Root         string `goftpd:"rootpath"`
	ShadowDB     string `goftpd:"shadow_db"`
	DefaultUser  string `goftpd:"default_user"`
	DefaultGroup string `goftpd:"default_group"`
	Hide         string `goftpd:"hide"`
	hideRE       *regexp.Regexp
}

func (f *FilesystemOpts) SetHideRE(r *regexp.Regexp) { f.hideRE = r }

type Filesystem struct {
	*FilesystemOpts
	chroot      billy.Filesystem
	shadow      Shadow
	permissions *acl.Permissions
	buffPool    sync.Pool
}

// NewFilesystem creates a new Filesystem with the given chroot (underlying fs) shadow (stores user/group meta data
// and permissions (check acl for paths, users and different scopes)
func NewFilesystem(opts *FilesystemOpts, chroot billy.Filesystem, shadow Shadow, permissions *acl.Permissions) (*Filesystem, error) {

	fs := Filesystem{
		FilesystemOpts: opts,
		chroot:         chroot,
		shadow:         shadow,
		permissions:    permissions,
		buffPool:       newBufferPoolWithSize(256 * 1024),
	}

	return &fs, nil
}

func (fs *Filesystem) GetBuffer() *[]byte {
	return fs.buffPool.Get().(*[]byte)
}

func (fs *Filesystem) PutBuffer(b *[]byte) {
	fs.buffPool.Put(b)
}

// Join tries to give back a safe path
func (fs Filesystem) Join(current string, params []string) string {

	path := strings.Join(params, " ")

	if !strings.HasPrefix(path, "/") {
		path = filepath.Join(current, path)
	}

	return path
}

// Stop closes any underlying resources
func (fs *Filesystem) Stop() error {
	if err := fs.shadow.Close(); err != nil {
		return err
	}
	return nil
}

// MakeDir checks to see if the user has permission to create a new directory. Does so if allowed
func (fs *Filesystem) MakeDir(path string, user *acl.User) error {
	if !fs.permissions.Match(acl.PermissionScopeMakeDir, path, user) {
		return acl.ErrPermissionDenied
	}

	// check for private
	if match, found := fs.permissions.MatchNoDefault(acl.PermissionScopePrivate, path, user); found && !match {
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

	if err := fs.shadow.Set(path, user.Name, user.PrimaryGroup); err != nil {
		return err
	}

	return nil
}

// DownloadFile checks to see if the user has permission to read the file (checking download
// permissions from high level to low level). Returns an io.ReadCloser if allowed
func (fs *Filesystem) DownloadFile(path string, user *acl.User) (ReadSeekCloser, int64, error) {
	if !fs.permissions.Match(acl.PermissionScopeDownload, path, user) {
		return nil, 0, acl.ErrPermissionDenied
	}

	// check for private
	if match, found := fs.permissions.MatchNoDefault(acl.PermissionScopePrivate, path, user); found && !match {
		return nil, 0, os.ErrNotExist
	}

	if fs.hideRE != nil {
		if fs.hideRE.MatchString(path) {
			// do not leak any information, just pretend
			// it doesnt exist
			return nil, 0, os.ErrNotExist
		}
	}

	f, err := fs.chroot.Open(path)
	if err != nil {
		return nil, 0, err
	}

	finfo, err := fs.chroot.Stat(path)
	if err != nil {
		return nil, 0, err
	}

	return f, finfo.Size(), nil
}

// UploadFile checks to see if the user has permission to write the file (checking upload
// permissions from high level to low level). Returns an io.Writer if allowed. Does not
// truncate a file
func (fs *Filesystem) UploadFile(path string, user *acl.User) (io.WriteCloser, error) {
	// TODO
	// need to check if we are currently uploading

	if !fs.permissions.Match(acl.PermissionScopeUpload, path, user) {
		return nil, acl.ErrPermissionDenied
	}

	// check for private
	if match, found := fs.permissions.MatchNoDefault(acl.PermissionScopePrivate, path, user); found && !match {
		return nil, acl.ErrPermissionDenied
	}

	// check if we would be able to delete it
	if !fs.permissions.Match(acl.PermissionScopeDelete, path, user) {

		// not allowed to globally delete, check if this is ours and we can delete our own
		if !fs.permissions.Match(acl.PermissionScopeDeleteOwn, path, user) {
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

	f, err := fs.chroot.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, defaultPerms)

	if err != nil {
		return nil, err
	}

	// wrap the file in our special Writer that allows us to manage the shadow fs
	writer := newWriteCloser(f, func() error {
		return fs.shadow.Set(path, user.Name, user.PrimaryGroup)
	})

	return writer, nil
}

// ResumeUploadFile checks to see if the user has permission to write the file (checking upload
// permissions from high level to low level). It also checks to see if they have resume writes.
// Returns an io.Writer if allowed.
func (fs *Filesystem) ResumeUploadFile(path string, user *acl.User) (io.WriteCloser, error) {
	if !fs.permissions.Match(acl.PermissionScopeUpload, path, user) {
		return nil, acl.ErrPermissionDenied
	}

	// check for private
	if match, found := fs.permissions.MatchNoDefault(acl.PermissionScopePrivate, path, user); found && !match {
		return nil, os.ErrNotExist
	}

	if fs.hideRE != nil {
		if fs.hideRE.MatchString(path) {
			// do not leak any information, just pretend
			// it doesnt exist
			return nil, os.ErrNotExist
		}
	}

	if !fs.permissions.Match(acl.PermissionScopeResume, path, user) {
		// not allowed to globally resume, check if this is ours and we can resume our own
		if !fs.permissions.Match(acl.PermissionScopeResumeOwn, path, user) {
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
		return fs.shadow.Set(path, user.Name, user.PrimaryGroup)
	})

	return writer, nil
}

// RenameFile checks to see if the user has permission to rename the file (checking rename and
// renameown scopes).
func (fs *Filesystem) RenameFile(oldpath, newpath string, user *acl.User) error {
	// make sure that the user has permission to upload to the new path
	if !fs.permissions.Match(acl.PermissionScopeUpload, newpath, user) {
		return acl.ErrPermissionDenied
	}

	// check for private
	if match, found := fs.permissions.MatchNoDefault(acl.PermissionScopePrivate, oldpath, user); found && !match {
		return os.ErrNotExist
	}
	if match, found := fs.permissions.MatchNoDefault(acl.PermissionScopePrivate, newpath, user); found && !match {
		return os.ErrNotExist
	}

	if fs.hideRE != nil {
		if fs.hideRE.MatchString(oldpath) || fs.hideRE.MatchString(newpath) {
			// do not leak any information, just pretend
			// it doesnt exist
			return os.ErrNotExist
		}
	}

	if !fs.permissions.Match(acl.PermissionScopeRename, oldpath, user) {

		// not allowed to globally rename, check if this is ours and we can rename our own
		if !fs.permissions.Match(acl.PermissionScopeRenameOwn, oldpath, user) {
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

	if oldpath == newpath {
		return errors.New("can not rename to self")
	}

	if err := fs.chroot.Rename(oldpath, newpath); err != nil {
		return err
	}

	if err := fs.shadow.Remove(oldpath); err != nil {
		return err
	}

	if err := fs.shadow.Set(newpath, user.Name, user.PrimaryGroup); err != nil {
		return err
	}

	return nil
}

// DeleteFile checks to see if the user has permission to delete the file (checking delete and
// deleteown scopes).
func (fs *Filesystem) DeleteFile(path string, user *acl.User) error {
	if !fs.permissions.Match(acl.PermissionScopeDelete, path, user) {

		// not allowed to globally delete, check if this is ours and we can delete our own
		if !fs.permissions.Match(acl.PermissionScopeDeleteOwn, path, user) {
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

	// check for private
	if match, found := fs.permissions.MatchNoDefault(acl.PermissionScopePrivate, path, user); found && !match {
		return os.ErrNotExist
	}

	if fs.hideRE != nil {
		if fs.hideRE.MatchString(path) {
			// do not leak any information, just pretend
			// it doesnt exist
			return os.ErrNotExist
		}
	}

	finfo, err := fs.chroot.Stat(path)
	if err != nil {
		return err
	}

	if finfo.IsDir() {
		return errors.New("can not delete directory.")
	}

	if err := fs.chroot.Remove(path); err != nil {
		return err
	}

	if err := fs.shadow.Remove(path); err != nil {
		return err
	}

	return nil
}

// DeleteDir checks to see if the user has permission to delete the dir (checking delete and
// deleteown scopes).
func (fs *Filesystem) DeleteDir(path string, user *acl.User) error {
	if !fs.permissions.Match(acl.PermissionScopeDelete, path, user) {

		// not allowed to globally delete, check if this is ours and we can delete our own
		if !fs.permissions.Match(acl.PermissionScopeDeleteOwn, path, user) {
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

	// check for private
	if match, found := fs.permissions.MatchNoDefault(acl.PermissionScopePrivate, path, user); found && !match {
		return os.ErrNotExist
	}

	finfo, err := fs.chroot.Stat(path)
	if err != nil {
		return err
	}

	if !finfo.IsDir() {
		return errors.New("can not delete file.")
	}

	if err := fs.chroot.Remove(path); err != nil {
		return err
	}

	if err := fs.shadow.Remove(path); err != nil {
		return err
	}

	return nil
}

// ListDir checks to see if the user has permission to list the dir and then does so.
// Has optimisation potential by being provided a FileList
func (fs *Filesystem) ListDir(path string, user *acl.User) (FileList, error) {
	if !fs.permissions.Match(acl.PermissionScopeDownload, path, user) {
		return nil, acl.ErrPermissionDenied
	}

	// check for private
	if match, found := fs.permissions.MatchNoDefault(acl.PermissionScopePrivate, path, user); found && !match {
		return nil, os.ErrNotExist
	}

	if fs.hideRE != nil {
		if fs.hideRE.MatchString(path) {
			// do not leak any information, just pretend
			// it doesnt exist
			return nil, os.ErrNotExist
		}
	}

	files, err := fs.chroot.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var results FileList

	for _, f := range files {
		fullpath := filepath.Join(path, f.Name())

		if fs.hideRE != nil {
			if fs.hideRE.MatchString(fullpath) {
				continue
			}
		}

		// check for private
		if match, found := fs.permissions.MatchNoDefault(acl.PermissionScopePrivate, fullpath, user); found && !match {
			continue
		}

		username, group, err := fs.shadow.Get(fullpath)
		if err != nil {
			username = fs.DefaultUser
			group = fs.DefaultGroup
		}

		// check if we have permission to see user and group, as it's hide, permissions are reversed
		if fs.permissions.Match(acl.PermissionScopeHideUser, fullpath, user) {
			username = fs.DefaultUser
		}
		if fs.permissions.Match(acl.PermissionScopeHideGroup, fullpath, user) {
			group = fs.DefaultGroup
		}

		results = append(results, FileInfo{
			FileInfo: f,
			Owner:    username,
			Group:    group,
		})
	}

	return results, nil
}

// checkOwnership checks to see if a user is an owner of a given path. Returns bool
// and an error
func (fs *Filesystem) checkOwnership(path string, user *acl.User) (bool, error) {
	username, _, err := fs.shadow.Get(path)
	if err != nil {
		return false, err
	}

	if username != strings.ToLower(user.Name) {
		return false, nil
	}

	return true, nil
}
