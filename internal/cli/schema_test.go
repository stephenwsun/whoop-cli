package cli

import (
	"encoding/json"
	"testing"
)

func TestCommandSchemaIsReadOnlyAndStable(t *testing.T) {
	document := commandSchema()
	if !document.ReadOnly || document.Version != 1 {
		t.Fatalf("unexpected schema metadata: %#v", document)
	}
	if len(document.Commands) < 10 || document.Commands[0].Name != "auth" {
		t.Fatalf("unexpected command list: %#v", document.Commands)
	}
	if _, err := json.Marshal(document); err != nil {
		t.Fatal(err)
	}
}

func TestValidatePolicyEnforcesReadOnlyAutomation(t *testing.T) {
	if err := validatePolicy("recovery", "", options{safetyProfile: "readonly", allowCommands: []string{"recovery"}}); err != nil {
		t.Fatal(err)
	}
	if err := validatePolicy("auth", "", options{safetyProfile: "readonly", noInput: true}); err == nil {
		t.Fatal("expected no-input auth rejection")
	}
	if err := validatePolicy("profile", "", options{safetyProfile: "readonly", allowCommands: []string{"recovery"}}); err == nil {
		t.Fatal("expected allowlist rejection")
	}
	if err := validatePolicy("profile", "", options{safetyProfile: "write"}); err == nil {
		t.Fatal("expected unsupported profile rejection")
	}
}
