package cmd

import (
	"log"

	"github.com/goftpd/goftpd/config"
	"github.com/spf13/cobra"
)

func init() {
	var cfg, username, password string

	var adduserCmd = &cobra.Command{
		Use:   "adduser",
		Short: "Check goftpd adduser",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := config.ParseFile(cfg)
			if err != nil {
				return err
			}

			// get auth
			auth, err := c.ParseAuthenticator()
			if err != nil {
				return err
			}

			// add user
			user, err := auth.AddUser(username, password)
			if err != nil {
				return err
			}

			log.Printf("created user '%s'", user.Name)

			return nil
		},
	}

	adduserCmd.Flags().StringVarP(&cfg, "config", "c", "goftpd.conf", "config file to load")
	adduserCmd.Flags().StringVarP(&username, "username", "u", "", "user to create")
	adduserCmd.Flags().StringVarP(&password, "password", "p", "", "password to add to user")

	adduserCmd.MarkFlagRequired("username")
	adduserCmd.MarkFlagRequired("password")

	rootCmd.AddCommand(adduserCmd)
}
