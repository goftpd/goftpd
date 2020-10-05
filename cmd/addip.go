package cmd

import (
	"log"

	"github.com/goftpd/goftpd/acl"
	"github.com/goftpd/goftpd/config"
	"github.com/spf13/cobra"
)

func init() {
	var cfg, username, mask string

	var addipCmd = &cobra.Command{
		Use:   "addip",
		Short: "Check goftpd addip",
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

			err = auth.UpdateUser(username, func(user *acl.User) error {
				return user.AddIP(mask)
			})
			if err != nil {
				return err
			}

			log.Printf("added mask '%s' to user '%s'", mask, username)

			return nil
		},
	}

	addipCmd.Flags().StringVarP(&cfg, "config", "c", "site/config/goftpd.conf", "config file to load")
	addipCmd.Flags().StringVarP(&username, "username", "u", "", "user to create")
	addipCmd.Flags().StringVarP(&mask, "mask", "m", "", "mask to add to user")

	addipCmd.MarkFlagRequired("username")
	addipCmd.MarkFlagRequired("mask")

	rootCmd.AddCommand(addipCmd)
}
