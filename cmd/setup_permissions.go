package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/chumam2050/ssh-connect/internal/config"
	"github.com/spf13/cobra"
)

func newSetupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "setup",
		Aliases: []string{"setup-permissions"},
		Short:   "Prepare ssh-connect config and SSH key permissions",
		Long:    "Fix .ssh permissions, ensure config exists, and auto-register accounts from *.pub keys.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if runtime.GOOS == "windows" {
				fmt.Println("setup is skipped on Windows")
				return nil
			}

			cfgExists := true
			if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
				cfgExists = false
			}
			cfg := loadConfigOrExit()

			home, _ := os.UserHomeDir()
			sshDir := filepath.Join(home, ".ssh")
			if err := os.MkdirAll(sshDir, 0o700); err != nil {
				return fmt.Errorf("failed to create %s: %w", sshDir, err)
			}
			if err := os.Chmod(sshDir, 0o700); err != nil {
				fmt.Printf("warning: failed to set %s to 0700: %v\n", sshDir, err)
			}

			permWarnings := 0
			if err := filepath.WalkDir(sshDir, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					permWarnings++
					fmt.Printf("warning: cannot access %s: %v\n", path, err)
					return nil
				}
				if d.IsDir() {
					if err := os.Chmod(path, 0o700); err != nil {
						permWarnings++
						fmt.Printf("warning: failed chmod 0700 for %s: %v\n", path, err)
					}
				}
				return nil
			}); err != nil {
				return fmt.Errorf("failed to walk %s: %w", sshDir, err)
			}

			scanned, scanWarnings, err := discoverAccountsFromPub(sshDir, home)
			if err != nil {
				return err
			}
			permWarnings += scanWarnings

			updatedAccounts := 0
			newAccounts := 0
			for _, acc := range scanned {
				idx := findAccountIndex(cfg, acc)
				if idx >= 0 {
					existing := cfg.Accounts[idx]
					cfg.Accounts[idx].Path = acc.Path
					if strings.TrimSpace(existing.Email) == "" {
						cfg.Accounts[idx].Email = acc.Email
					}
					if strings.TrimSpace(existing.Name) == "" {
						cfg.Accounts[idx].Name = acc.Name
					}
					updatedAccounts++
					continue
				}

				cfg.Accounts = append(cfg.Accounts, acc)
				newAccounts++
			}

			if !cfgExists {
				fmt.Printf("created config file: %s\n", cfgPath)
			}
			saveConfigOrExit(cfg)

			fmt.Printf("done: %d account(s) added, %d account(s) updated, %d warning(s)\n", newAccounts, updatedAccounts, permWarnings)
			return nil
		},
	}

	return cmd
}

func discoverAccountsFromPub(sshDir, home string) ([]config.Account, int, error) {
	accounts := make([]config.Account, 0)
	warnings := 0
	seenName := map[string]struct{}{}

	err := filepath.WalkDir(sshDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			warnings++
			fmt.Printf("warning: cannot read %s: %v\n", path, err)
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".pub") {
			return nil
		}

		privPath := strings.TrimSuffix(path, ".pub")
		info, statErr := os.Stat(privPath)
		if statErr != nil || info.IsDir() {
			warnings++
			fmt.Printf("warning: private key pair not found for %s\n", path)
			return nil
		}

		if err := os.Chmod(privPath, 0o600); err != nil {
			warnings++
			fmt.Printf("warning: failed chmod 0600 for %s: %v\n", privPath, err)
		} else {
			fmt.Printf("updated: %s (0600)\n", privPath)
		}

		identifier := readPublicKeyIdentifier(path)
		email, name := deriveIdentity(identifier, d.Name())

		for {
			if _, exists := seenName[name]; !exists {
				break
			}
			name = name + "-key"
		}
		seenName[name] = struct{}{}

		accounts = append(accounts, config.Account{
			Name:       name,
			Email:      email,
			Path:       collapseHome(privPath, home),
			Passphrase: "",
		})
		return nil
	})
	if err != nil {
		return nil, warnings, fmt.Errorf("scan %s: %w", sshDir, err)
	}

	sort.Slice(accounts, func(i, j int) bool {
		return accounts[i].Name < accounts[j].Name
	})
	return accounts, warnings, nil
}

func readPublicKeyIdentifier(pubPath string) string {
	f, err := os.Open(pubPath)
	if err != nil {
		return ""
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	if !s.Scan() {
		return ""
	}
	parts := strings.Fields(strings.TrimSpace(s.Text()))
	if len(parts) < 3 {
		return ""
	}
	return parts[2]
}

func deriveIdentity(identifier, fileName string) (email string, name string) {
	identifier = strings.TrimSpace(identifier)
	if strings.Contains(identifier, "@") {
		email = identifier
		user := strings.SplitN(identifier, "@", 2)[0]
		name = sanitizeName(user)
		if name != "" {
			return email, name
		}
	}

	name = sanitizeName(identifier)
	if name != "" {
		return email, name
	}

	base := strings.TrimSuffix(fileName, ".pub")
	return email, sanitizeName(base)
}

func sanitizeName(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" {
		return "account"
	}

	b := strings.Builder{}
	for _, r := range v {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
			continue
		}
		if r == ' ' {
			b.WriteRune('-')
		}
	}
	out := strings.Trim(b.String(), "-.")
	if out == "" {
		return "account"
	}
	return out
}

func collapseHome(path, home string) string {
	cleanHome := filepath.Clean(home)
	cleanPath := filepath.Clean(path)
	if cleanHome != "" && (cleanPath == cleanHome || strings.HasPrefix(cleanPath, cleanHome+string(os.PathSeparator))) {
		rel := strings.TrimPrefix(cleanPath, cleanHome)
		rel = strings.TrimPrefix(rel, string(os.PathSeparator))
		if rel == "" {
			return "~"
		}
		return "~/" + rel
	}
	return cleanPath
}

func findAccountIndex(cfg *config.Config, candidate config.Account) int {
	if cfg == nil {
		return -1
	}
	for i, acc := range cfg.Accounts {
		if strings.TrimSpace(candidate.Email) != "" && strings.EqualFold(strings.TrimSpace(acc.Email), strings.TrimSpace(candidate.Email)) {
			return i
		}
	}
	for i, acc := range cfg.Accounts {
		if strings.EqualFold(strings.TrimSpace(acc.Name), strings.TrimSpace(candidate.Name)) {
			return i
		}
	}
	for i, acc := range cfg.Accounts {
		if config.ExpandHome(strings.TrimSpace(acc.Path)) == config.ExpandHome(strings.TrimSpace(candidate.Path)) {
			return i
		}
	}
	return -1
}
