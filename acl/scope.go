package acl

// PermissionScope is an "enum" for the different filesystem permissions
// supported.
type PermissionScope string

const (
	PermissionScopeDownload  PermissionScope = "download"
	PermissionScopeUpload                    = "upload"
	PermissionScopeRename                    = "rename"
	PermissionScopeRenameOwn                 = "renameown"
	PermissionScopeDelete                    = "delete"
	PermissionScopeDeleteOwn                 = "deleteown"
	PermissionScopeResume                    = "resume"
	PermissionScopeResumeOwn                 = "resumeown"
	PermissionScopeMakeDir                   = "makedir"
	PermissionScopeShowUser                  = "showuser"
	PermissionScopeShowGroup                 = "showgroup"
	PermissionScopePrivate                   = "private"
)

var StringToPermissionScope = map[string]PermissionScope{
	string(PermissionScopeDownload):  PermissionScopeDownload,
	string(PermissionScopeUpload):    PermissionScopeUpload,
	string(PermissionScopeRename):    PermissionScopeRename,
	string(PermissionScopeRenameOwn): PermissionScopeRenameOwn,
	string(PermissionScopeDelete):    PermissionScopeDelete,
	string(PermissionScopeDeleteOwn): PermissionScopeDeleteOwn,
	string(PermissionScopeResume):    PermissionScopeResume,
	string(PermissionScopeResumeOwn): PermissionScopeResumeOwn,
	string(PermissionScopeMakeDir):   PermissionScopeMakeDir,
	string(PermissionScopeShowUser):  PermissionScopeShowUser,
	string(PermissionScopeShowGroup): PermissionScopeShowGroup,
	string(PermissionScopePrivate):   PermissionScopePrivate,
}
