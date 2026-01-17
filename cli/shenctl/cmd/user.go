/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/elmq0022/shen/cli/shenctl/cmd/client"
	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users in Shen",
	Long: `Manage users in Shen.

Users can have one of three roles:
  - admin:   Full access to manage all resources
  - user:    Can manage own PATs and view own groups
  - service: Service account with token-only access (cannot login to Shen)

Examples:
  shenctl user list
  shenctl user create alice admin
  shenctl user update alice --role user
  shenctl user delete alice`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("user called")
	},
}

var listUserCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users",
	Long: `List active users in Shen.

Displays username and role for each active user.
By default, only the first 10 users are shown. Use --all to retrieve the complete list.
Requires admin privileges.

Examples:
  shenctl user list
  shenctl user list --all`,
	Run: func(cmd *cobra.Command, args []string) {
		all, _ := cmd.Flags().GetBool("all")

		cursor := ""
		limit := 10

		for {
			users, err := client.ListActiveUsers(cursor, limit)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not read api: %v", err)
				return
			}

			for _, user := range users {
				fmt.Printf("%s\t%s\n", user.Username, user.Role)
			}

			if !all || len(users) < limit {
				break
			}

			cursor = users[len(users)-1].Username
		}
	},
}

var createUserCmd = &cobra.Command{
	Use:   "create <username> <role>",
	Short: "Create a new user",
	Long: `Create a new user in Shen.

The role must be one of: admin, user, or service.
You will be prompted to enter a password for admin and user roles.
Service accounts do not have passwords and authenticate via tokens only.

Requires admin privileges.

Examples:
  shenctl user create alice admin
  shenctl user create bob user
  shenctl user create ci-deploy service`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("user called")
	},
}

var updateUserCmd = &cobra.Command{
	Use:   "update <username>",
	Short: "Update a user",
	Long: `Update an existing user in Shen.

You can update the user's role and/or password.
  - Only admins can change user roles
  - Users can change their own password
  - Admins can change any user's password

Examples:
  shenctl user update alice --role user
  shenctl user update alice --password
  shenctl user update alice --role admin --password`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var deleteUserCmd = &cobra.Command{
	Use:   "delete <username>",
	Short: "Delete a user",
	Long: `Delete a user from Shen.

This performs a soft delete, marking the user as inactive.
The user will no longer be able to authenticate.

Requires admin privileges.

Examples:
  shenctl user delete alice`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("user called")
	},
}

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(listUserCmd)
	userCmd.AddCommand(createUserCmd)
	userCmd.AddCommand(updateUserCmd)
	userCmd.AddCommand(deleteUserCmd)

	// list user flags
	listUserCmd.Flags().BoolP("all", "a", false, "use to retrive a complete list of user instead of the first 10")

	// update user flags
	updateUserCmd.Flags().StringP("password", "p", "", "New password (prompts if flag is present but empty)")
	updateUserCmd.Flags().StringP("role", "r", "", "New role for the user (admin, user, or service)")
}
