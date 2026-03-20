package cmd

import (
	"fmt"

	"github.com/chumam2050/ssh-connect/internal/config"
	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List accounts and servers",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := loadConfigOrExit()
			printAccounts(cfg)
			fmt.Println()
			printServers(cfg)
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "account",
		Short: "List available accounts",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := loadConfigOrExit()
			printAccounts(cfg)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "server",
		Short: "List available servers",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := loadConfigOrExit()
			printServers(cfg)
		},
	})

	return cmd
}

func printAccounts(cfg *config.Config) {
	if len(cfg.Accounts) == 0 {
		fmt.Println("accounts: (empty)")
		return
	}
	fmt.Println("accounts:")
	for _, a := range cfg.Accounts {
		fmt.Printf("  - %s | %s | %s\n", a.Name, a.Email, a.Path)
	}
}

func printServers(cfg *config.Config) {
	if len(cfg.Servers) == 0 {
		fmt.Println("servers: (empty)")
		return
	}
	fmt.Println("servers:")
	for _, s := range cfg.Servers {
		fmt.Printf("  - %s | %s@%s:%d | default_account=%s\n", s.Name, s.Username, s.Host, s.Port, s.DefaultAccount)
	}
}
