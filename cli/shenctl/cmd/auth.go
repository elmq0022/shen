/*
Copyright © 2026 Aaron Elmquist
*/
package cmd

import (
	"github.com/elmq0022/shen/cli/shenctl/cmd/client"
	"github.com/spf13/cobra"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
	Long:  `Manage authentication with the shen server. Use login to authenticate and logout to clear credentials.`,
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to shen",
	Long:  `Authenticate with the shen server and store credentials locally.`,
	Run: func(cmd *cobra.Command, args []string) {
		client.Login()
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from shen",
	Long:  `Clear stored credentials and logout from the shen server.`,
	Run: func(cmd *cobra.Command, args []string) {
		client.Logout()
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(loginCmd, logoutCmd)
}
