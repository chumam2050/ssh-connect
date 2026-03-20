package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/chumam2050/ssh-connect/internal/config"
	sshrun "github.com/chumam2050/ssh-connect/internal/ssh"
	"github.com/spf13/cobra"
)

func newToCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "to [servername]",
		Short: "Connect to server name",
		Args:  cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			cfg, err := config.Load(cfgPath)
			if err != nil || cfg == nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			suggestions := make([]string, 0, len(cfg.Servers))
			needle := strings.ToLower(strings.TrimSpace(toComplete))
			for _, s := range cfg.Servers {
				name := strings.TrimSpace(s.Name)
				if name == "" {
					continue
				}
				if needle == "" || strings.HasPrefix(strings.ToLower(name), needle) {
					desc := fmt.Sprintf("%s@%s:%d | account=%s", s.Username, s.Host, s.Port, s.DefaultAccount)
					suggestions = append(suggestions, name+"\t"+desc)
				}
			}

			return suggestions, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			serverName := args[0]
			cfg := loadConfigOrExit()

			server, _ := cfg.FindServer(serverName)
			if server == nil {
				return fmt.Errorf("server %q not found", serverName)
			}

			account, _ := cfg.FindAccount(server.DefaultAccount)
			if account == nil {
				return fmt.Errorf("default account %q for server %q not found", server.DefaultAccount, server.Name)
			}

			keyPath := config.ExpandHome(account.Path)
			if _, err := os.Stat(keyPath); err != nil {
				return fmt.Errorf("ssh key path %q is not accessible: %w", keyPath, err)
			}

			fmt.Printf("connecting to %s (%s@%s:%d) using account %s\n", server.Name, server.Username, server.Host, server.Port, account.Name)
			if err := sshrun.Connect(server.Username, server.Host, server.Port, keyPath, account.Passphrase); err != nil {
				return fmt.Errorf("ssh failed: %w", err)
			}
			return nil
		},
	}
	return cmd
}
