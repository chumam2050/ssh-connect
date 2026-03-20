package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func Connect(username, host string, port int, keyPath string, passphrase string) error {
	if strings.TrimSpace(passphrase) != "" {
		agentEnv, cleanup, err := startSSHAgent()
		if err != nil {
			return err
		}
		defer cleanup()

		if err := addKey(agentEnv, keyPath, passphrase); err != nil {
			return err
		}
		return runSSH(agentEnv, username, host, port, keyPath)
	}

	return runSSH(nil, username, host, port, keyPath)
}

func runSSH(extraEnv []string, username, host string, port int, keyPath string) error {
	args := []string{}
	if keyPath != "" {
		args = append(args, "-i", keyPath)
	}
	if port > 0 {
		args = append(args, "-p", strconv.Itoa(port))
	}
	args = append(args, fmt.Sprintf("%s@%s", username, host))

	cmd := exec.Command("ssh", args...)
	if len(extraEnv) > 0 {
		cmd.Env = append(os.Environ(), extraEnv...)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func startSSHAgent() ([]string, func(), error) {
	out, err := exec.Command("ssh-agent", "-s").CombinedOutput()
	if err != nil {
		return nil, func() {}, fmt.Errorf("failed to start ssh-agent: %v (%s)", err, strings.TrimSpace(string(out)))
	}

	parsed := parseAgentEnv(string(out))
	if len(parsed) == 0 {
		return nil, func() {}, fmt.Errorf("failed to parse ssh-agent environment")
	}

	cleanup := func() {
		cmd := exec.Command("ssh-agent", "-k")
		cmd.Env = append(os.Environ(), parsed...)
		_ = cmd.Run()
	}

	return parsed, cleanup, nil
}

func addKey(agentEnv []string, keyPath string, passphrase string) error {
	cmd := exec.Command("ssh-add", keyPath)
	cmd.Env = append(os.Environ(), agentEnv...)
	cmd.Stdin = strings.NewReader(passphrase + "\n")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to unlock private key: %v (%s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func parseAgentEnv(out string) []string {
	result := make([]string, 0, 2)
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "SSH_AUTH_SOCK=") || strings.HasPrefix(line, "SSH_AGENT_PID=") {
			part := strings.SplitN(line, ";", 2)[0]
			if strings.Contains(part, "=") {
				result = append(result, part)
			}
		}
	}
	return result
}
