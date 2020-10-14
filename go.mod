module github.com/goftpd/goftpd

go 1.15

require (
	github.com/DataDog/zstd v1.4.5 // indirect
	github.com/alexedwards/argon2id v0.0.0-20200802152012-2464efd3196b
	github.com/dgraph-io/badger v1.6.2
	github.com/dgraph-io/badger/v2 v2.0.1-rc1.0.20201007220711-3b5f17cee813
	github.com/go-git/go-billy/v5 v5.0.0
	github.com/gobwas/glob v0.2.3
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/golang/snappy v0.0.2 // indirect
	github.com/google/go-cmp v0.5.0
	github.com/oragono/go-ident v0.0.0-20200511222032-830550b1d775
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v0.0.5
	github.com/vadv/gopher-lua-libs v0.1.1
	github.com/vmihailenco/msgpack/v5 v5.0.0-beta.1
	github.com/yuin/gopher-lua v0.0.0-20200816102855-ee81675732da
	golang.org/x/net v0.0.0-20201006153459-a7d1128ccaa0 // indirect
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208
	golang.org/x/sys v0.0.0-20201007165808-a893ed343c85 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	layeh.com/gopher-luar v1.0.8
)

replace github.com/go-git/go-billy/v5 => github.com/jawr/go-billy/v5 v5.0.1-0.20200914114554-78517ac908a2
