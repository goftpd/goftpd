package cmd

import (
	"context"
	"log"

	"github.com/goftpd/goftpd/config"
	"github.com/goftpd/goftpd/ftp"
	"github.com/spf13/cobra"
)

func init() {
	var configPath string

	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run goftpd",
		RunE: func(cmd *cobra.Command, args []string) error {

			cfg, err := config.ParseFile(configPath)
			if err != nil {
				return err
			}

			serverOpts, err := cfg.ParseServerOpts()
			if err != nil {
				return err
			}

			fs, err := cfg.ParseFS()
			if err != nil {
				return err
			}
			defer fs.Stop()

			// get auth
			auth, err := cfg.ParseAuthenticator()
			if err != nil {
				return err
			}

			// get script engine
			se, err := cfg.ParseScripts()
			if err != nil {
				return err
			}

			server, err := ftp.NewServer(serverOpts, fs, auth, se)
			if err != nil {
				return err
			}

			ctx := context.Background()

			log.Printf("listen and serve..")

			if err := server.ListenAndServe(ctx); err != nil {
				return err
			}

			return nil
		},
	}

	runCmd.Flags().StringVarP(&configPath, "config", "c", "site/config/goftpd.conf", "config file to load")

	rootCmd.AddCommand(runCmd)
}
