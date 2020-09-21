package cmd

import (
	"log"

	"github.com/goftpd/goftpd/config"
	"github.com/spf13/cobra"
)

func init() {
	var cfg string

	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "Check goftpd config",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := config.ParseFile(cfg)
			if err != nil {
				return err
			}

			serverOpts, err := c.ParseServerOpts()
			if err != nil {
				return err
			}

			log.Printf("%+v", serverOpts)

			return nil
		},
	}

	configCmd.Flags().StringVarP(&cfg, "config", "c", "goftpd.conf", "config file to load")

	rootCmd.AddCommand(configCmd)
}
