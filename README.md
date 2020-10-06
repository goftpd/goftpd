[![Coverage Status](https://coveralls.io/repos/github/goftpd/goftpd/badge.svg?branch=master)](https://coveralls.io/github/goftpd/goftpd?branch=master)
[![GoDoc](https://godoc.org/github.com/goftpd/goftpd?status.svg)](https://godoc.org/github.com/goftpd/goftpd)

### GOFTPD
Aim is to be a replacement for glftpd, but with less bloat, more testing and be
more extendable. 

A working config:

```
# variables
# ---------
# you can set variables that will be reused in the script, there are some
# caveats; variables must be preceded by a space, i.e. $var would work, 
# but !$var would not.
var admins -user =admin
var special =group1 =group2

# acl settings
# ------------
# the glob syntax is:
# * matches matches any sequence of non-Separator characters
# ** matches any sequence of characters, including Seperator
# ? matches any single non-Separator character
# [cHaRs] matches any range of charecters in "cHaRs"

# permissions always default to false, matches are always from largest to smallest
# in terms of pattern length, and first pattern to match wins, no more are
# checked. this means that there is no need to append '!*' to rules

# acl download also includes the ability to list
acl download 	/**		*
acl upload 		/**		*
acl resume 		/**		*
acl resumeown	/**		*
acl delete		/**		*
acl deleteown	/**		*
acl resume		/**		*
acl resumeown	/**		*
acl show_user 	/**		*
acl show_group	/**		*

acl show_user 	/requests/** $admins
acl show_group	/requests/** $admins

# special makedir rules
acl makedir		/mp3/????-??-??/*/*		*
acl makedir 	/mp3/* $admins

# an example of an admin setup
acl private /admin $admins
acl private /admin/** $admins

# an example pre setup
acl private /groups $admins $special
acl private /groups/** $admins $special

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
# permissions are to show user/group use this user/group
fs default_user		nobody
fs default_group	ohhai

# regexp. show these from listing and prevent from being downloaded
# also protects them from rename and delete
fs show (?i)\.(message)$
```

## To Run
Make sure you have Go installed (https://golang.org/dl/). Download this repo,
i.e. with git:

`git clone https://github.com/goftpd/goftpd.git`

or wget:

`wget https://github.com/goftpd/goftpd/archive/master.zip && unzip master`

If you don't want to edit the conf, then:

`mkdir site/data && mkdir site/config`

Create some self signed certs (feature to autogen this will be added):

`openssl req -x509 -newkey rsa:4096 -keyout site/config/key.pem -out site/config/cert.pem -days 365 -nodes`

Then run it:

`go run main.go adduser -u goftpd -p ohemgeedontusethis`
`go run main.go run`

Congratulations, you are now a hacker.


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
acl upload /path** -user =group !*
acl download /path** -user =group *
acl rename /path** -user =group !*
acl renameown /path** -user =group *
acl delete /path** -user =group !*
acl deleteown /path** -user =group *
acl resume /path** -user =group !*
acl resumeown /path** -user =group *
acl makedir /path** -user =group !*
acl list /path** -user =group *
acl showuser /** !=staff *
acl showgroup /** !=staff *
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
script post 'RETR' event site/scripts/wowow.lua
script pre 'MKD' trigger site/scripts/omg.lua
script post 'SITE BOOP' trigger site/scripts/doaboop.lua
```
