package cmd

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/chumam2050/ssh-connect/internal/config"
	"github.com/spf13/cobra"
)

func newAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add account or server",
	}

	cmd.AddCommand(newAddAccountCommand())
	cmd.AddCommand(newAddServerCommand())

	return cmd
}

func newAddAccountCommand() *cobra.Command {
	var (
		name       string
		email      string
		path       string
		passphrase string
	)

	cmd := &cobra.Command{
		Use:   "account",
		Short: "Add a new SSH account",
		RunE: func(cmd *cobra.Command, args []string) error {
			reader := bufio.NewReader(os.Stdin)
			interactive, _ := cmd.Flags().GetBool("interactive")
			if interactive || name == "" || email == "" || path == "" {
				var err error
				if name, err = ask(reader, "Account name", name); err != nil {
					return err
				}
				if email, err = ask(reader, "Account email", email); err != nil {
					return err
				}
				if path, err = ask(reader, "SSH key path", path); err != nil {
					return err
				}
			}

			if strings.TrimSpace(name) == "" || strings.TrimSpace(path) == "" {
				return fmt.Errorf("name and path are required")
			}

			cfg := loadConfigOrExit()
			if found, _ := cfg.FindAccount(name); found != nil {
				return fmt.Errorf("account %q already exists", name)
			}

			cfg.Accounts = append(cfg.Accounts, config.Account{
				Name:       strings.TrimSpace(name),
				Email:      strings.TrimSpace(email),
				Path:       strings.TrimSpace(path),
				Passphrase: passphrase,
			})
			saveConfigOrExit(cfg)
			fmt.Printf("account %q added\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "account name")
	cmd.Flags().StringVar(&email, "email", "", "account email")
	cmd.Flags().StringVar(&path, "path", "", "ssh identity key path")
	cmd.Flags().StringVar(&passphrase, "passphrase", "", "private key passphrase")
	cmd.Flags().BoolP("interactive", "i", false, "force interactive prompts")

	return cmd
}

func newAddServerCommand() *cobra.Command {
	var (
		name           string
		username       string
		host           string
		port           int
		defaultAccount string
	)

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Add a new server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := loadConfigOrExit()
			accountNames := getAccountNames(cfg)
			if len(accountNames) == 0 {
				return fmt.Errorf("no account available, add account first")
			}

			reader := bufio.NewReader(os.Stdin)
			interactive, _ := cmd.Flags().GetBool("interactive")
			if interactive || name == "" || username == "" || host == "" || defaultAccount == "" {
				var err error
				if name, err = ask(reader, "Server name", name); err != nil {
					return err
				}
				if username, err = ask(reader, "SSH username", username); err != nil {
					return err
				}
				if host, err = ask(reader, "SSH host", host); err != nil {
					return err
				}
				if port, err = askInt(reader, "SSH port", port); err != nil {
					return err
				}
				if defaultAccount, err = askValidAccount(reader, defaultAccount, accountNames); err != nil {
					return err
				}
			}

			if strings.TrimSpace(name) == "" || strings.TrimSpace(host) == "" || strings.TrimSpace(defaultAccount) == "" {
				return fmt.Errorf("name, host, and default account are required")
			}
			if port <= 0 {
				port = 22
			}

			if found, _ := cfg.FindServer(name); found != nil {
				return fmt.Errorf("server %q already exists", name)
			}
			if found, _ := cfg.FindAccount(defaultAccount); found == nil {
				return fmt.Errorf("default account %q does not exist", defaultAccount)
			}

			cfg.Servers = append(cfg.Servers, config.Server{
				Name:           strings.TrimSpace(name),
				Username:       strings.TrimSpace(username),
				Host:           strings.TrimSpace(host),
				Port:           port,
				DefaultAccount: strings.TrimSpace(defaultAccount),
			})
			saveConfigOrExit(cfg)
			fmt.Printf("server %q added\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "server name")
	cmd.Flags().StringVar(&username, "username", "", "ssh username")
	cmd.Flags().StringVar(&host, "host", "", "hostname or ip")
	cmd.Flags().IntVar(&port, "port", 22, "ssh port")
	cmd.Flags().StringVar(&defaultAccount, "account", "", "default account name")
	cmd.Flags().BoolP("interactive", "i", false, "force interactive prompts")

	_ = cmd.RegisterFlagCompletionFunc("account", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		names := getAccountNames(cfg)
		if strings.TrimSpace(toComplete) == "" {
			return names, cobra.ShellCompDirectiveNoFileComp
		}
		matches := make([]string, 0)
		needle := strings.ToLower(strings.TrimSpace(toComplete))
		for _, n := range names {
			if strings.HasPrefix(strings.ToLower(n), needle) {
				matches = append(matches, n)
			}
		}
		return matches, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func getAccountNames(cfg *config.Config) []string {
	if cfg == nil {
		return nil
	}
	names := make([]string, 0, len(cfg.Accounts))
	for _, a := range cfg.Accounts {
		if strings.TrimSpace(a.Name) != "" {
			names = append(names, a.Name)
		}
	}
	sort.Strings(names)
	return names
}

func askValidAccount(reader *bufio.Reader, current string, accounts []string) (string, error) {
	if len(accounts) == 0 {
		return "", fmt.Errorf("account list is empty")
	}
	fmt.Printf("Available accounts: %s\n", strings.Join(accounts, ", "))

	for {
		value, err := ask(reader, "Default account", current)
		if err != nil {
			return "", err
		}
		value = strings.TrimSpace(value)
		if value == "" {
			fmt.Println("default account is required")
			continue
		}

		if exact, ok := findExactAccount(value, accounts); ok {
			return exact, nil
		}

		matches := findAccountPrefixMatches(value, accounts)
		if len(matches) == 1 {
			fmt.Printf("using account: %s\n", matches[0])
			return matches[0], nil
		}
		if len(matches) > 1 {
			fmt.Printf("account not specific. matches: %s\n", strings.Join(matches, ", "))
			current = ""
			continue
		}

		fmt.Printf("account %q not found. try again.\n", value)
		current = ""
	}
}

func findExactAccount(value string, accounts []string) (string, bool) {
	needle := strings.TrimSpace(strings.ToLower(value))
	for _, a := range accounts {
		if strings.ToLower(a) == needle {
			return a, true
		}
	}
	return "", false
}

func findAccountPrefixMatches(value string, accounts []string) []string {
	needle := strings.TrimSpace(strings.ToLower(value))
	result := make([]string, 0)
	for _, a := range accounts {
		if strings.HasPrefix(strings.ToLower(a), needle) {
			result = append(result, a)
		}
	}
	return result
}
