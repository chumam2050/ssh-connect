package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newPassphraseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "passphrase",
		Short: "Manage SSH account passphrase",
	}
	cmd.AddCommand(newPassphraseAccountCommand())
	return cmd
}

func newPassphraseAccountCommand() *cobra.Command {
	var (
		name       string
		value      string
		clearValue bool
	)

	cmd := &cobra.Command{
		Use:   "account [name]",
		Short: "Set passphrase for an account",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := loadConfigOrExit()
			if len(args) == 1 {
				name = strings.TrimSpace(args[0])
			}
			if name == "" {
				choices := make([]string, 0, len(cfg.Accounts))
				for _, a := range cfg.Accounts {
					choices = append(choices, a.Name)
				}
				picked, err := selectFromList("Select account:", choices)
				if err != nil {
					return err
				}
				name = picked
			}

			_, idx := cfg.FindAccount(name)
			if idx < 0 {
				return fmt.Errorf("account %q not found", name)
			}

			if clearValue {
				cfg.Accounts[idx].Passphrase = ""
				saveConfigOrExit(cfg)
				fmt.Printf("passphrase for account %q cleared\n", name)
				return nil
			}

			if value == "" {
				reader := bufio.NewReader(os.Stdin)
				input, err := ask(reader, "Passphrase (leave empty to clear)", "")
				if err != nil {
					return err
				}
				value = input
			}

			cfg.Accounts[idx].Passphrase = value
			saveConfigOrExit(cfg)
			if value == "" {
				fmt.Printf("passphrase for account %q cleared\n", name)
			} else {
				fmt.Printf("passphrase for account %q updated\n", name)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "account name")
	cmd.Flags().StringVar(&value, "value", "", "passphrase value")
	cmd.Flags().BoolVar(&clearValue, "clear", false, "clear passphrase")
	return cmd
}
