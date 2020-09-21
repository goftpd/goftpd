package cmd

import (
	"context"

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

			server, err := ftp.NewServer(serverOpts, fs)
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

	runCmd.Flags().StringVarP(&configPath, "config", "c", "goftpd.conf", "config file to load")

	rootCmd.AddCommand(runCmd)
}
