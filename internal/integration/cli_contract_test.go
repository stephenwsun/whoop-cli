package integration

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type cliResult struct {
	stdout string
	stderr string
	err    error
}

func buildCLI(t *testing.T) string {
	t.Helper()
	name := "whoop"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binary := filepath.Join(t.TempDir(), name)
	cmd := exec.Command("go", "build", "-trimpath", "-o", binary, "../../cmd/whoop")
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, output)
	}
	return binary
}

func runCLI(binary string, args ...string) cliResult {
	cmd := exec.Command(binary, args...)
	cmd.Env = append(os.Environ(), "WHOOP_CLIENT_ID=", "WHOOP_CLIENT_SECRET=")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return cliResult{stdout: stdout.String(), stderr: stderr.String(), err: err}
}

func TestCLIContractWithoutCredentials(t *testing.T) {
	binary := buildCLI(t)

	version := runCLI(binary, "--version")
	if version.err != nil || !strings.Contains(version.stdout, "dev") || version.stderr != "" {
		t.Fatalf("--version = %#v", version)
	}

	versionJSON := runCLI(binary, "version", "--json")
	var versionDocument map[string]string
	if versionJSON.err != nil || json.Unmarshal([]byte(versionJSON.stdout), &versionDocument) != nil || versionDocument["version"] == "" || versionJSON.stderr != "" {
		t.Fatalf("version --json = %#v", versionJSON)
	}

	schema := runCLI(binary, "schema", "--json")
	var schemaDocument struct {
		Version  int  `json:"version"`
		ReadOnly bool `json:"read_only"`
		Commands []struct {
			Name string `json:"name"`
		} `json:"commands"`
	}
	if schema.err != nil || json.Unmarshal([]byte(schema.stdout), &schemaDocument) != nil || schemaDocument.Version != 1 || !schemaDocument.ReadOnly || schema.stderr != "" {
		t.Fatalf("schema --json = %#v", schema)
	}
	foundVersion := false
	for _, command := range schemaDocument.Commands {
		if command.Name == "version" {
			foundVersion = true
		}
	}
	if !foundVersion {
		t.Fatalf("schema does not advertise version: %#v", schemaDocument.Commands)
	}

	diagnose := runCLI(binary, "config", "diagnose", "--json")
	var diagnosis map[string]any
	if diagnose.err != nil || json.Unmarshal([]byte(diagnose.stdout), &diagnosis) != nil || diagnose.stderr != "" {
		t.Fatalf("config diagnose --json = %#v", diagnose)
	}

	badMode := runCLI(binary, "config", "diagnose", "--json", "--plain")
	if exitCode(badMode.err) != 2 || !strings.Contains(badMode.stderr, "mutually exclusive") || badMode.stdout != "" {
		t.Fatalf("mutually exclusive modes = %#v", badMode)
	}

	noInput := runCLI(binary, "auth", "--no-input")
	if exitCode(noInput.err) != 2 || !strings.Contains(noInput.stderr, "requires browser input") || noInput.stdout != "" {
		t.Fatalf("auth --no-input = %#v", noInput)
	}
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	if exitError, ok := err.(*exec.ExitError); ok {
		return exitError.ExitCode()
	}
	return -1
}
