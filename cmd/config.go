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

			if _, err := c.ParseServerOpts(); err != nil {
				return err
			}

			if _, err := c.ParseFS(); err != nil {
				return err
			}

			log.Println("config file parsed ok")

			return nil
		},
	}

	configCmd.Flags().StringVarP(&cfg, "config", "c", "goftpd.conf", "config file to load")

	rootCmd.AddCommand(configCmd)
}
