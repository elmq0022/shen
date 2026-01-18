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

var applicationCmd = &cobra.Command{
	Use:   "application",
	Short: "Manage applications in Shen",
	Long: `Manage applications in Shen.

Applications represent external systems that can be integrated with Shen
for role-based access control.

Examples:
  shenctl application list
  shenctl application create myapp
  shenctl application delete myapp`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("application called")
	},
}

var listApplicationCmd = &cobra.Command{
	Use:   "list",
	Short: "List all applications",
	Long: `List active applications in Shen.

Displays the name of each active application.
By default, only the first 10 applications are shown. Use --all to retrieve the complete list.
Supports cursor-based pagination with --cursor and --limit flags.
Requires admin privileges.

Examples:
  shenctl application list
  shenctl application list --all
  shenctl application list --limit 5
  shenctl application list --cursor "myapp" --limit 10`,
	Run: func(cmd *cobra.Command, args []string) {
		all, _ := cmd.Flags().GetBool("all")
		cursor, _ := cmd.Flags().GetString("cursor")
		limit, _ := cmd.Flags().GetInt("limit")

		for {
			apps, err := client.ListActiveApplications(cursor, limit)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not read api: %v\n", err)
				os.Exit(1)
			}

			for _, app := range apps {
				fmt.Println(app.Name)
			}

			if !all || len(apps) < limit {
				break
			}

			cursor = apps[len(apps)-1].Name
		}
	},
}

var createApplicationCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new application",
	Long: `Create a new application in Shen.

Application names are automatically normalized to lowercase.

Requires admin privileges.

Examples:
  shenctl application create myapp
  shenctl application create MyApp    # stored as "myapp"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		app, err := client.CreateApplication(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating application: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("%s\n", app.Name)
	},
}

var deleteApplicationCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete an application",
	Long: `Delete an application from Shen.

This performs a soft delete, marking the application as inactive.

Requires admin privileges.

Examples:
  shenctl application delete myapp`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if err := client.DeleteApplication(name); err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting application: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Application %s deleted successfully\n", name)
	},
}

func init() {
	rootCmd.AddCommand(applicationCmd)
	applicationCmd.AddCommand(listApplicationCmd)
	applicationCmd.AddCommand(createApplicationCmd)
	applicationCmd.AddCommand(deleteApplicationCmd)

	// list application flags
	listApplicationCmd.Flags().BoolP("all", "a", false, "retrieve a complete list of applications instead of the first 10")
	listApplicationCmd.Flags().StringP("cursor", "c", "", "cursor for pagination (application name to start after)")
	listApplicationCmd.Flags().IntP("limit", "l", 10, "number of applications to retrieve per request")
}
