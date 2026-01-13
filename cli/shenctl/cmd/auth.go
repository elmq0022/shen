/*
Copyright © 2026 Aaron Elmquist
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/elmq0022/shen/cli/shenctl/cmd/client"
	"github.com/elmq0022/shen/cli/shenctl/utils"
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
		var password string
		var username string
		var err error

		username, _ = cmd.Flags().GetString("username")
		if username == "" {
			username, err = utils.ReadUsername()
			if err != nil {
				fmt.Printf("failed to read username: %v\n", err)
				os.Exit(1)
			}
		}

		password, _ = cmd.Flags().GetString("password")
		if password == "" {
			password, err = utils.ReadPassword()
			if err != nil {
				fmt.Printf("failed to read password: %v\n", err)
				os.Exit(1)
			}
		}

		if err := client.Login(username, password); err != nil {
			fmt.Println("login failed")
			os.Exit(1)
		}

		fmt.Println("login successful")
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from shen",
	Long:  `Clear stored credentials and logout from the shen server.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := client.Logout(); err != nil {
			fmt.Printf("logout failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("logout successful")
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(loginCmd, logoutCmd)

	// Add flags for login command
	loginCmd.Flags().StringP("username", "u", "", "Username for authentication (skips interactive prompt)")
	loginCmd.Flags().StringP("password", "p", "", "Password for authentication (skips interactive prompt)")
}
