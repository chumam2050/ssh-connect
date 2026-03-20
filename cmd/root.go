package cmd

import (
	"fmt"
	"os"

	"github.com/chumam2050/ssh-connect/internal/config"
	"github.com/spf13/cobra"
)

var cfgPath string

var rootCmd = &cobra.Command{
	Use:   "ssh-connect",
	Short: "Manage multiple SSH accounts and servers",
	Long:  "ssh-connect helps you manage SSH account identities and named servers from one config file.",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", config.DefaultPath(), "path to ssh-connect config JSON")
	rootCmd.AddCommand(newToCommand())
	rootCmd.AddCommand(newListCommand())
	rootCmd.AddCommand(newAddCommand())
	rootCmd.AddCommand(newDeleteCommand())
	rootCmd.AddCommand(newPassphraseCommand())
	rootCmd.AddCommand(newSetupCommand())
}

func loadConfigOrExit() *config.Config {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config %s: %v\n", cfgPath, err)
		os.Exit(1)
	}
	return cfg
}

func saveConfigOrExit(cfg *config.Config) {
	if err := config.Save(cfgPath, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "failed to save config %s: %v\n", cfgPath, err)
		os.Exit(1)
	}
}
