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

var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "Manage groups in Shen",
	Long: `Manage groups in Shen.

Groups are organizational units that can have users as members.
Groups can be assigned roles for applications to enable RBAC.

Examples:
  shenctl group list
  shenctl group create engineering
  shenctl group delete engineering
  shenctl group add-users engineering alice bob
  shenctl group remove-users engineering alice`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("group called")
	},
}

var listGroupCmd = &cobra.Command{
	Use:   "list",
	Short: "List all groups",
	Long: `List active groups in Shen.

Displays the name of each active group.
By default, only the first 10 groups are shown. Use --all to retrieve the complete list.
Supports cursor-based pagination with --cursor and --limit flags.
Requires admin privileges.

Examples:
  shenctl group list
  shenctl group list --all
  shenctl group list --limit 5
  shenctl group list --cursor "engineering" --limit 10`,
	Run: func(cmd *cobra.Command, args []string) {
		all, _ := cmd.Flags().GetBool("all")
		cursor, _ := cmd.Flags().GetString("cursor")
		limit, _ := cmd.Flags().GetInt("limit")

		for {
			groups, err := client.ListActiveGroups(cursor, limit)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not read api: %v\n", err)
				os.Exit(1)
			}

			for _, group := range groups {
				fmt.Println(group.Name)
			}

			if !all || len(groups) < limit {
				break
			}

			cursor = groups[len(groups)-1].Name
		}
	},
}

var createGroupCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new group",
	Long: `Create a new group in Shen.

Group names are automatically normalized to lowercase.

Requires admin privileges.

Examples:
  shenctl group create engineering
  shenctl group create Engineering    # stored as "engineering"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		group, err := client.CreateGroup(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating group: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("%s\n", group.Name)
	},
}

var deleteGroupCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a group",
	Long: `Delete a group from Shen.

This performs a soft delete, marking the group as inactive.
All group memberships will be removed (cascade delete).

Requires admin privileges.

Examples:
  shenctl group delete engineering`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		if err := client.DeleteGroup(name); err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting group: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Group %s deleted successfully\n", name)
	},
}

var addUsersCmd = &cobra.Command{
	Use:   "add-users <group> <user1> [user2] ...",
	Short: "Add users to a group",
	Long: `Add one or more users to a group.

Requires admin privileges (or group manager privileges in future versions).

Examples:
  shenctl group add-users engineering alice
  shenctl group add-users engineering alice bob charlie`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]
		usernames := args[1:]

		result, err := client.AddUsersToGroup(groupName, usernames)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error adding users to group: %v\n", err)
			os.Exit(1)
		}

		if len(result.Added) > 0 {
			fmt.Printf("Added to %s: %v\n", groupName, result.Added)
		}
		if len(result.NotFound) > 0 {
			fmt.Fprintf(os.Stderr, "Users not found: %v\n", result.NotFound)
		}
	},
}

var removeUsersCmd = &cobra.Command{
	Use:   "remove-users <group> <user1> [user2] ...",
	Short: "Remove users from a group",
	Long: `Remove one or more users from a group.

Requires admin privileges (or group manager privileges in future versions).

Examples:
  shenctl group remove-users engineering alice
  shenctl group remove-users engineering alice bob`,
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]
		usernames := args[1:]

		result, err := client.RemoveUsersFromGroup(groupName, usernames)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error removing users from group: %v\n", err)
			os.Exit(1)
		}

		if len(result.Removed) > 0 {
			fmt.Printf("Removed from %s: %v\n", groupName, result.Removed)
		}
		if len(result.NotFound) > 0 {
			fmt.Fprintf(os.Stderr, "Users not found: %v\n", result.NotFound)
		}
	},
}

var addRoleCmd = &cobra.Command{
	Use:   "add-role <group> <application> <role>",
	Short: "Assign a role to a group for an application",
	Long: `Assign an application role to a group.

This enables RBAC by mapping groups to application roles.
Available roles: authenticated, viewer, auditor, operator, admin

Requires admin privileges.

Examples:
  shenctl group add-role engineering myapp viewer
  shenctl group add-role ops myapp operator`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]
		application := args[1]
		role := args[2]

		if err := client.AddRoleToGroup(groupName, application, role); err != nil {
			fmt.Fprintf(os.Stderr, "Error adding role to group: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Role %s for %s added to group %s\n", role, application, groupName)
	},
}

var removeRoleCmd = &cobra.Command{
	Use:   "remove-role <group> <application> <role>",
	Short: "Remove a role from a group for an application",
	Long: `Remove an application role from a group.

Requires admin privileges.

Examples:
  shenctl group remove-role engineering myapp viewer
  shenctl group remove-role ops myapp operator`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]
		application := args[1]
		role := args[2]

		if err := client.RemoveRoleFromGroup(groupName, application, role); err != nil {
			fmt.Fprintf(os.Stderr, "Error removing role from group: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Role %s for %s removed from group %s\n", role, application, groupName)
	},
}

var listRolesCmd = &cobra.Command{
	Use:   "list-roles <group> [application]",
	Short: "List roles assigned to a group",
	Long: `List roles assigned to a group, optionally filtered by application.

Displays the application and role for each assignment.
By default, only the first 10 roles are shown. Use --all to retrieve the complete list.
Supports cursor-based pagination with --cursor-app, --cursor-role, and --limit flags.

Requires admin privileges.

Examples:
  shenctl group list-roles engineering
  shenctl group list-roles engineering myapp
  shenctl group list-roles engineering --all
  shenctl group list-roles engineering --limit 5`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		groupName := args[0]
		application := ""
		if len(args) > 1 {
			application = args[1]
		}

		all, _ := cmd.Flags().GetBool("all")
		cursorApp, _ := cmd.Flags().GetString("cursor-app")
		cursorRole, _ := cmd.Flags().GetString("cursor-role")
		limit, _ := cmd.Flags().GetInt("limit")

		for {
			roles, err := client.ListGroupRoles(groupName, application, cursorApp, cursorRole, limit)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error listing group roles: %v\n", err)
				os.Exit(1)
			}

			for _, role := range roles {
				fmt.Printf("%s\t%s\n", role.Application, role.Role)
			}

			if !all || len(roles) < limit {
				break
			}

			// Update cursors for next page
			lastRole := roles[len(roles)-1]
			cursorApp = lastRole.Application
			cursorRole = lastRole.Role
		}
	},
}

func init() {
	rootCmd.AddCommand(groupCmd)
	groupCmd.AddCommand(listGroupCmd)
	groupCmd.AddCommand(createGroupCmd)
	groupCmd.AddCommand(deleteGroupCmd)
	groupCmd.AddCommand(addUsersCmd)
	groupCmd.AddCommand(removeUsersCmd)
	groupCmd.AddCommand(addRoleCmd)
	groupCmd.AddCommand(removeRoleCmd)
	groupCmd.AddCommand(listRolesCmd)

	// list group flags
	listGroupCmd.Flags().BoolP("all", "a", false, "retrieve a complete list of groups instead of the first 10")
	listGroupCmd.Flags().StringP("cursor", "c", "", "cursor for pagination (group name to start after)")
	listGroupCmd.Flags().IntP("limit", "l", 10, "number of groups to retrieve per request")

	// list-roles flags
	listRolesCmd.Flags().BoolP("all", "a", false, "retrieve a complete list of roles instead of the first 10")
	listRolesCmd.Flags().String("cursor-app", "", "cursor for pagination (application name)")
	listRolesCmd.Flags().String("cursor-role", "", "cursor for pagination (role name)")
	listRolesCmd.Flags().IntP("limit", "l", 10, "number of roles to retrieve per request")
}
