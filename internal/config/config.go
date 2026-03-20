package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Account struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Path       string `json:"path"`
	Passphrase string `json:"passphrase"`
}

type Server struct {
	Name           string `json:"name"`
	Username       string `json:"username"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	DefaultAccount string `json:"default_account"`
}

type Config struct {
	Accounts []Account `json:"accounts"`
	Servers  []Server  `json:"servers"`
}

func DefaultPath() string {
	if env := strings.TrimSpace(os.Getenv("SSH_CONNECT_CONFIG")); env != "" {
		return env
	}
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return "ssh-connect-config.json"
	}
	return filepath.Join(home, ".ssh-connect", "ssh-connect-config.json")
}

func Load(path string) (*Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, err
	}

	cfg := &Config{}
	if len(strings.TrimSpace(string(content))) == 0 {
		return cfg, nil
	}
	if err := json.Unmarshal(content, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

func Save(path string, cfg *Config) error {
	if cfg == nil {
		return errors.New("nil config")
	}
	data, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	data = append(data, '\n')

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
		return fmt.Errorf("ensure config dir: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func (c *Config) FindAccount(name string) (*Account, int) {
	for i := range c.Accounts {
		if c.Accounts[i].Name == name {
			return &c.Accounts[i], i
		}
	}
	return nil, -1
}

func (c *Config) FindServer(name string) (*Server, int) {
	for i := range c.Servers {
		if c.Servers[i].Name == name {
			return &c.Servers[i], i
		}
	}
	return nil, -1
}

func ExpandHome(path string) string {
	if path == "~" {
		home, _ := os.UserHomeDir()
		return home
	}
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		if home != "" {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
