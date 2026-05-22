package sftp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

const agentDialTimeout = 5 * time.Second

var (
	ErrEncryptedKeyNoPassphrase = errors.New("SSH key is encrypted but no key_pass provided")
)

type AuthConfig struct {
	password   string
	privateKey string
	keyPass    string
	useAgent   bool
}

func ParseAuthFromURL(u *url.URL) AuthConfig {
	cfg := AuthConfig{
		password:   "",
		privateKey: "",
		keyPass:    "",
		useAgent:   false,
	}

	if u.User != nil {
		if pw, ok := u.User.Password(); ok {
			cfg.password = pw
		}
	}

	cfg.privateKey = u.Query().Get("key")
	cfg.keyPass = u.Query().Get("key_pass")
	cfg.useAgent = u.Query().Get("agent") == "true"

	return cfg
}

func (a AuthConfig) BuildAuthMethods() ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod

	if a.useAgent && os.Getenv("SSH_AUTH_SOCK") != "" {
		agentSigners, agentErr := a.buildAgentSigners()
		if agentErr != nil {
			return nil, agentErr
		}
		if len(agentSigners) > 0 {
			methods = append(methods, ssh.PublicKeys(agentSigners...))
		}
	}

	keyPath := a.privateKey
	if keyPath == "" && len(methods) == 0 && a.password == "" {
		keyPath = findDefaultKeyPath()
	}

	if keyPath != "" {
		signer, keyErr := parseKeyAtPath(keyPath, a.keyPass)
		if keyErr != nil {
			return nil, keyErr
		}
		methods = append(methods, ssh.PublicKeys(signer))
	}

	if a.password != "" {
		methods = append(methods, ssh.Password(a.password))
	}

	return methods, nil
}

func (a AuthConfig) buildAgentSigners() ([]ssh.Signer, error) {
	sock := os.Getenv("SSH_AUTH_SOCK")

	dialer := &net.Dialer{Timeout: agentDialTimeout} //nolint:exhaustruct // only Timeout is relevant for agent socket
	conn, dialErr := dialer.DialContext(context.Background(), "unix", sock)
	if dialErr != nil {
		return nil, fmt.Errorf("failed to connect to SSH agent: %w", dialErr)
	}
	defer conn.Close()

	agentClient := agent.NewClient(conn)
	signers, signErr := agentClient.Signers()
	if signErr != nil {
		return nil, fmt.Errorf("failed to get signers from SSH agent: %w", signErr)
	}

	return signers, nil
}

func parseKeyAtPath(path, keyPass string) (ssh.Signer, error) {
	path = expandTilde(path)

	keyBytes, readErr := os.ReadFile(path)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read SSH key %q: %w", path, readErr)
	}

	if keyPass != "" {
		signer, parseErr := ssh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(keyPass))
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse encrypted SSH key %q: %w", path, parseErr)
		}
		return signer, nil
	}

	signer, parseErr := ssh.ParsePrivateKey(keyBytes)
	if parseErr != nil && isEncryptedKeyError(parseErr) {
		return nil, fmt.Errorf("SSH key %q: %w", path, ErrEncryptedKeyNoPassphrase)
	}
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse SSH key %q: %w", path, parseErr)
	}

	return signer, nil
}

func isEncryptedKeyError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "password protected") ||
		strings.Contains(msg, "encrypted")
}

func getDefaultKeyNames() []string {
	return []string{"id_ed25519", "id_ecdsa", "id_rsa"}
}

func findDefaultKeyPath() string {
	home, homeErr := os.UserHomeDir()
	if homeErr != nil {
		return ""
	}

	sshDir := filepath.Join(home, ".ssh")

	for _, name := range getDefaultKeyNames() {
		candidate := filepath.Join(sshDir, name)
		if _, statErr := os.Stat(candidate); statErr == nil {
			return candidate
		}
	}

	return ""
}

func expandTilde(p string) string {
	if p == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return p
		}
		return home
	}

	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return p
		}
		return filepath.Join(home, p[2:])
	}

	return p
}
