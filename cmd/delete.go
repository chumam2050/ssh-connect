package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete account or server",
	}

	cmd.AddCommand(newDeleteAccountCommand())
	cmd.AddCommand(newDeleteServerCommand())

	return cmd
}

func newDeleteAccountCommand() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "account [name]",
		Short: "Delete SSH account",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := loadConfigOrExit()
			if len(args) == 1 {
				name = args[0]
			}
			if name == "" {
				choices := make([]string, 0, len(cfg.Accounts))
				for _, a := range cfg.Accounts {
					choices = append(choices, a.Name)
				}
				picked, err := selectFromList("Select account to delete:", choices)
				if err != nil {
					return err
				}
				name = picked
			}

			_, idx := cfg.FindAccount(name)
			if idx < 0 {
				return fmt.Errorf("account %q not found", name)
			}

			for _, s := range cfg.Servers {
				if s.DefaultAccount == name {
					return fmt.Errorf("account %q is used by server %q", name, s.Name)
				}
			}

			cfg.Accounts = append(cfg.Accounts[:idx], cfg.Accounts[idx+1:]...)
			saveConfigOrExit(cfg)
			fmt.Printf("account %q deleted\n", name)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "account name")
	return cmd
}

func newDeleteServerCommand() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "server [name]",
		Short: "Delete server",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := loadConfigOrExit()
			if len(args) == 1 {
				name = args[0]
			}
			if name == "" {
				choices := make([]string, 0, len(cfg.Servers))
				for _, s := range cfg.Servers {
					choices = append(choices, s.Name)
				}
				picked, err := selectFromList("Select server to delete:", choices)
				if err != nil {
					return err
				}
				name = picked
			}

			_, idx := cfg.FindServer(name)
			if idx < 0 {
				return fmt.Errorf("server %q not found", name)
			}

			cfg.Servers = append(cfg.Servers[:idx], cfg.Servers[idx+1:]...)
			saveConfigOrExit(cfg)
			fmt.Printf("server %q deleted\n", name)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "server name")
	return cmd
}
