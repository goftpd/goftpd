[![Coverage Status](https://coveralls.io/repos/github/goftpd/goftpd/badge.svg?branch=master)](https://coveralls.io/github/goftpd/goftpd?branch=master)
[![GoDoc](https://godoc.org/github.com/goftpd/goftpd?status.svg)](https://godoc.org/github.com/goftpd/goftpd)

### GOFTPD
Aim is to be a replacement for glftpd, but with less bloat, more testing and be
more extendable. 

## Config
Please look in `site/config/goftpd.conf` for a working config with lots of
comments and an overview of current features.

## To Run
Make sure you have Go installed (https://golang.org/dl/). Download this repo,
i.e. with git:

`git clone https://github.com/goftpd/goftpd.git`

or wget:

`wget https://github.com/goftpd/goftpd/archive/master.zip && unzip master`

If you don't want to edit the conf, then:

`mkdir site/data && mkdir site/config`

Then run it (change 127.0.0.1 for your host):

```
go run main.go gencert -h 127.0.0.1
go run main.go adduser -u goftpd -p ohemgeedontusethis
go run main.go addip -u goftpd -m *@*.*.*.*
go run main.go run
```

Congratulations, you are now a hacker.

## PZS-NG
Install PZS-NG:

```
./configure --disable-glftpd-specific
```

Your `zsconfig.h` should use absolute paths, i.e.:

```
#define sitepath_dir                 "/srv/goftpd/site/data"
#define group_dirs                   "/srv/goftpd/site/data/groups/"
#define zip_dirs                     "/srv/goftpd/site/data/zip"
#define sfv_dirs                     "/srv/goftpd/site/data/sfv"

#define log                          "/srv/goftpd/site/ftp-data/logs/glftpd.log"
#define storage                      "/srv/goftpd/site/ftp-data/pzs-ng/"
```

And of course ensure it's create: `mkdir /srv/goftpd/site/ftp-data/`. Run `make`
and then `cp zipscript/src/zipscript-c /srv/goftpd/site/bin/`


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

## Issues
Currently there are some issues for 32bit based systems. Once resolved in
dependencies we should be able to compile.
