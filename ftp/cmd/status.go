package cmd

type Status struct {
	Code    int
	Message string
}

var (
	StatusOK                      Status = Status{200, "Command OK."}
	StatusSuperfluous                    = Status{202, "Command not implemented, superfluous at this site."}
	StatusServiceReady                   = Status{220, "Service ready for new user."}
	StatusCommandUnrecognised            = Status{500, "Syntax error, command unrecognized."}
	StatusSyntaxError                    = Status{501, "Syntax error in parameters or arguments."}
	StatusNotImplemented                 = Status{502, "Command not implemented."}
	StatusBadCommandSequence             = Status{503, "Bad sequence of commands."}
	StatusParameterNotImplemented        = Status{504, "Command not implemented for that parameter."}
	StatusSystemStatus                   = Status{211, "System status, or system help reply."}
	StatusDirectoryStatus                = Status{212, "Directory status."}
	StatusFileStatus                     = Status{213, "File status."}
	StatusHelpMessage                    = Status{214, "Help message."}
	StatusSystemType                     = Status{215, "NAME system type."}
	StatusClosingControl                 = Status{221, "Service closing control connection."}
	StatusServiceUnavailable             = Status{421, "Service not available, closing control connection."}
	StatusDataAlreadyOpen                = Status{125, "Data connection already open; transfer starting."}
	StatusDataOpenNoTransfer             = Status{225, "Data connection open; no transfer in progress."}
	StatusCantOpenDataConnection         = Status{425, "Can't open data connection."}
	StatusDataClosedOK                   = Status{226, "Closing data connection. Requested file action successful."}
	StatusBadProtectionLevel             = Status{534, "Protection Level '%s' is not accepted."}
	StatusDataCloseAborted               = Status{426, "Connection closed; transfer aborted."}
	StatusPassiveMode                    = Status{227, "Entering Passive Mode. %s"}
	StatusLongPassiveMode                = Status{228, "Entering Long Passive Mode (long address, port)."}
	StatusExtendedPassiveMode            = Status{229, "Entering Extended Passive Mode (|||port|)."}
	StatusUserLoggedIn                   = Status{230, "User '%s' logged in, proceed."}
	StatusUserLoggedOut                  = Status{232, "Logout command noted, will complete when transfer done."}
	StatusSecurityExchangeOK             = Status{234, "Authentication mechanism accepted."}
	StatusNotLoggedIn                    = Status{530, "Not logged in."}
	StatusNeedPassword                   = Status{331, "User name okay, need password."}
	StatusNeedAccount                    = Status{332, "Need account for login."}
	StatusNeedAccountToStor              = Status{532, "Need account for storing files."}
	StatusTransferStatusOK               = Status{150, "File status okay; about to open data connection."}
	StatusFileActionOK                   = Status{250, "Requested file action okay, completed."}
	StatusPathCreated                    = Status{257, `"%s" created.`}
	StatusPendingMoreInfo                = Status{350, "Requested file action pending further information."}
	StatusActionNotOK                    = Status{550, "Requested action not taken."}
	StatusActionAbortedError             = Status{451, "Requested action aborted. Local error in processing."}
	StatusPageTypeUnknown                = Status{551, "Requested action aborted. Page type unknown."}
	StatusNoDiskFree                     = Status{452, "Requested action not taken. Insufficient storage space in system. File unavailable (e.g., file busy)."}
	StatusBadFilename                    = Status{553, "Requested action not taken. File name not allowed."}
	StatusPermissionDenied               = Status{550, "Permission denied"}
)
