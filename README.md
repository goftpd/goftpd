[![Coverage Status](https://coveralls.io/repos/github/goftpd/goftpd/badge.svg?branch=master)](https://coveralls.io/github/goftpd/goftpd?branch=master)
[![GoDoc](https://godoc.org/github.com/goftpd/goftpd?status.svg)](https://godoc.org/github.com/goftpd/goftpd)

### GOFTPD
Aim is to be a replacement for glftpd, but with less bloat, more testing and be
more extendable. 

A working config:

```
# acl settings
# ------------
acl download 	/ 		*
acl upload 		/ 		*
acl resume 		/ 		*
acl hideuser 	/		!-admin *
acl hidegroup	/		!-admin *
acl list 		/		*
acl makedir		/		*

# server settings
# ---------------
server sitename_short 	go
server sitename_long 	goftpd
server host				::
server port				2121
# range of passive ports allowed  to be used
server passive_ports	1000 5000
# used for pasv
server public_ip		127.0.0.1
# required unless tls_autogen (TODO)
server tls_cert_file	site/cert.pem
server tls_key_file		site/key.pem

# fs based 
# --------
fs rootpath			site/data
# optional path where shadow fs database will be kept
fs shadow_db		shadow.db
# default_* if user or group isnt found in shadowdb or 
# permissions are to hide user/group use this user/group
fs default_user		nobody
fs default_group	ohhai

# regexp. hide these from listing and prevent from being downloaded
# also protects them from rename and delete
fs hide (?i)\.(message)$
```

## Ramblings
The core will implement the FTP RFC with pluggable Auth and Filesystem
components. This means that in the future, if someone were crazy, they could
authenticate users with Facebook and have the underlying storage in S3.

Config will be adjusted slightly, currently thinking to keep it similar to
glftpd, but with some namespacing, this will be seen through examples below.

ACL is definied in a similar way as glftpd, but with no flags. Flags are
essentially the same thing as groups and I'm yet to hear an argument to keep
them around. Currently implemented ACL Filesystem scopes are:

```
acl upload /path -user =group !*
acl download /path -user =group *
acl rename /path -user =group !*
acl renameown /path -user =group *
acl delete /path -user =group !*
acl deleteown /path -user =group *
acl resume /path -user =group !*
acl resumeown /path -user =group *
acl makedir /path -user =group !*
acl list /path -user =group *
acl hideuser / !=staff *
acl hidegroup / !=staff *
```

The filesystem currently does not use UID/GID as a way of storing meta data.
Instead we use a shadow filesystem which is essentially a key value store where
the key is a hash of the lowercased path with the value being the owner's
username and primary group. This would allow the FTPD to be run on all platforms
that GO can be compiled on (*looks at windows*), but this isn't a primary motive
or intended feature. Feedback/critique on this particular design is most
welcome.

There will be a scripting engine and lots of hooks. The scripting 
API is yet to be decided, and the language choices are lua or javascript 
(alternative suggestions welcome). These hooks will be in the form of an event 
(asynchronous) or a trigger (synchronous). Some examples might look like:

```
script post 'RETR' event scripts/wowow.lua
script pre 'MKD' trigger scripts/omg.lua
script post 'SITE BOOP' trigger scripts/doaboop.lua
```
