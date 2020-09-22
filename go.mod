module github.com/goftpd/goftpd

go 1.15

require (
	github.com/dgraph-io/badger v1.6.2
	github.com/dgraph-io/badger/v2 v2.2007.2
	github.com/go-git/go-billy/v5 v5.0.0
	github.com/gobwas/glob v0.2.3
	github.com/jawr/go-billy v3.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/segmentio/fasthash v1.0.3
	github.com/spf13/cobra v0.0.5
	github.com/yargevad/filepathx v0.0.0-20161019152617-907099cb5a62
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208
	gopkg.in/src-d/go-billy.v3 v3.1.0 // indirect
)

replace github.com/go-git/go-billy/v5 => github.com/jawr/go-billy/v5 v5.0.1-0.20200914114554-78517ac908a2
