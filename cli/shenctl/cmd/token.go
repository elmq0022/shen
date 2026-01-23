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

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage personal access tokens (PATs)",
	Long: `Manage personal access tokens (PATs) in Shen.

PATs are used to authenticate with applications integrated with Shen.
Tokens can only be viewed once when created - store them securely.

Examples:
  shenctl token list
  shenctl token list --user alice              # Admin only
  shenctl token create my-token myapp`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("token called")
	},
}

var listTokenCmd = &cobra.Command{
	Use:   "list",
	Short: "List your tokens",
	Long: `List personal access tokens.

By default, lists your own tokens. Admins can list another user's tokens
with the --user flag.

Examples:
  shenctl token list
  shenctl token list --all
  shenctl token list --user alice      # Admin only
  shenctl token list --limit 5 --cursor 10`,
	Run: func(cmd *cobra.Command, args []string) {
		all, _ := cmd.Flags().GetBool("all")
		user, _ := cmd.Flags().GetString("user")
		cursorStr, _ := cmd.Flags().GetString("cursor")
		limit, _ := cmd.Flags().GetInt("limit")

		var cursor int32
		if cursorStr != "" {
			var c int
			_, err := fmt.Sscanf(cursorStr, "%d", &c)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: invalid cursor value: %v\n", err)
				os.Exit(1)
			}
			cursor = int32(c)
		}

		for {
			tokens, err := client.ListTokens(user, cursor, limit)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			for _, token := range tokens {
				fmt.Printf("%d\t%s\t%s\t%s\n",
					token.ID,
					token.Name,
					token.ApplicationName,
					token.ExpiresAt.Time.Format("2006-01-02"),
				)
			}

			if !all || len(tokens) < limit {
				break
			}

			cursor = tokens[len(tokens)-1].ID
		}
	},
}

var createTokenCmd = &cobra.Command{
	Use:   "create <token-name> <application>",
	Short: "Create a new personal access token",
	Long: `Create a new personal access token for an application.

The token is displayed only once - store it securely.

Examples:
  shenctl token create my-ci-token myapp
  shenctl token create deploy-token myapp --exp 2026-04-01T00:00:00Z`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		application := args[1]
		expiration, _ := cmd.Flags().GetString("exp")

		patResp, err := client.CreateToken(name, application, expiration)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(patResp.PAT)
	},
}

func init() {
	rootCmd.AddCommand(tokenCmd)
	tokenCmd.AddCommand(listTokenCmd)
	tokenCmd.AddCommand(createTokenCmd)

	// list token flags
	listTokenCmd.Flags().BoolP("all", "a", false, "retrieve all tokens instead of first 10")
	listTokenCmd.Flags().StringP("user", "u", "", "list tokens for specific user (admin only)")
	listTokenCmd.Flags().StringP("cursor", "c", "", "cursor for pagination (token ID)")
	listTokenCmd.Flags().IntP("limit", "l", 10, "number of tokens per request")

	// create token flags
	createTokenCmd.Flags().StringP("exp", "e", "", "expiration time in ISO 8601 format (default: 30 days)")
}
