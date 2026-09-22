package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var ErrNotFound = errors.New("credential not found")

type Store interface {
	Get(context.Context, string) (string, error)
	Set(context.Context, string, string) error
}

type KeychainStore struct{ Service string }

func (s KeychainStore) service() string {
	if s.Service != "" {
		return s.Service
	}
	return "whoop-cli"
}
func (s KeychainStore) Get(ctx context.Context, account string) (string, error) {
	cmd := exec.CommandContext(ctx, "security", "find-generic-password", "-s", s.service(), "-a", account, "-w")
	out, err := cmd.Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("read Keychain: %w", err)
	}
	value := strings.TrimSpace(string(out))
	if value == "" {
		return "", ErrNotFound
	}
	return value, nil
}
func (s KeychainStore) Set(ctx context.Context, account, value string) error {
	cmd := exec.CommandContext(ctx, "security", "add-generic-password", "-U", "-s", s.service(), "-a", account, "-w", value)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("write Keychain: %s", redactCommandError(err, out))
	}
	return nil
}

// FileStore is a 0600 fallback for non-macOS systems and deterministic tests.
type FileStore struct{ Path string }

func (s FileStore) path() string {
	if s.Path != "" {
		return s.Path
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "whoop", "credentials.json")
}
func (s FileStore) Get(ctx context.Context, key string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	data, err := os.ReadFile(s.path())
	if errors.Is(err, os.ErrNotExist) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("read credential file: %w", err)
	}
	var values map[string]string
	if err := json.Unmarshal(data, &values); err != nil {
		return "", fmt.Errorf("decode credential file: %w", err)
	}
	value, ok := values[key]
	if !ok || value == "" {
		return "", ErrNotFound
	}
	return value, nil
}
func (s FileStore) Set(ctx context.Context, key, value string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	path := s.path()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create credential directory: %w", err)
	}
	values := map[string]string{}
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &values); err != nil {
			return fmt.Errorf("decode credential file: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read credential file: %w", err)
	}
	values[key] = value
	data, err := json.Marshal(values)
	if err != nil {
		return fmt.Errorf("encode credential file: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".credentials-*")
	if err != nil {
		return fmt.Errorf("create credential file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write credential file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("install credential file: %w", err)
	}
	return nil
}

func DefaultStore() Store {
	if runtime.GOOS == "darwin" {
		return KeychainStore{Service: "whoop-cli"}
	}
	return FileStore{}
}

func redactCommandError(err error, output []byte) string {
	if len(output) > 120 {
		output = output[:120]
	}
	return strings.TrimSpace(fmt.Sprintf("%v (%s)", err, string(output)))
}
func RandomState() (string, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
