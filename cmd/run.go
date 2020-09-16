package cmd

import (
	"context"

	"github.com/dgraph-io/badger/v2"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/goftpd/goftpd/acl"
	"github.com/goftpd/goftpd/ftp"
	"github.com/goftpd/goftpd/vfs"
	"github.com/spf13/cobra"
)

func init() {
	var config string

	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run goftpd",
		RunE: func(cmd *cobra.Command, args []string) error {

			ufs := osfs.New("site/")

			opt := badger.DefaultOptions("shadow.db")

			db, err := badger.Open(opt)
			if err != nil {
				return err
			}
			defer db.Close()

			shadowFS := vfs.NewShadowStore(db)

			lines := []string{
				"download / *",
				"upload / *",
				"makedir / *",
				"list / *",
			}

			var rules []acl.Rule
			for _, l := range lines {
				r, err := acl.NewRule(l)
				if err != nil {
					return err
				}
				rules = append(rules, r)
			}

			perms, err := acl.NewPermissions(rules)
			if err != nil {
				return err
			}

			fs, err := vfs.NewFilesystem(ufs, shadowFS, perms)
			if err != nil {
				return err
			}

			opts := ftp.ServerOpts{
				Name:        "goftpd",
				Port:        2121,
				Host:        "::",
				PublicIP:    "172.20.5.192",
				TLSCertFile: "cert.pem",
				TLSKeyFile:  "key.pem",
			}

			server, err := ftp.NewServer(opts, fs)
			if err != nil {
				return err
			}

			ctx := context.Background()

			if err := server.ListenAndServe(ctx); err != nil {
				return err
			}

			return nil
		},
	}

	runCmd.Flags().StringVarP(&config, "config", "c", "goftpd.conf", "config file to load")

	rootCmd.AddCommand(runCmd)
}
