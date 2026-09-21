package cli

import (
	"os"

	"github.com/stephensun/whoop-cli/internal/output"
)

type schemaDocument struct {
	Name      string         `json:"name"`
	Version   int            `json:"version"`
	ReadOnly  bool           `json:"read_only"`
	ExitCodes []exitCodeSpec `json:"exit_codes"`
	Global    []flagSpec     `json:"global_flags"`
	Commands  []commandSpec  `json:"commands"`
}

type exitCodeSpec struct {
	Code    int    `json:"code"`
	Meaning string `json:"meaning"`
}

type flagSpec struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description"`
}

type commandSpec struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	ReadOnly    bool       `json:"read_only"`
	Flags       []flagSpec `json:"flags,omitempty"`
}

func commandSchema() schemaDocument {
	return schemaDocument{
		Name:     "whoop",
		Version:  1,
		ReadOnly: true,
		ExitCodes: []exitCodeSpec{
			{Code: 0, Meaning: "success"},
			{Code: 1, Meaning: "API, OAuth, or runtime failure"},
			{Code: 2, Meaning: "usage or local configuration failure"},
		},
		Global: []flagSpec{
			{Name: "json", Type: "bool", Description: "emit stable JSON"},
			{Name: "plain", Type: "bool", Description: "emit stable tab-separated output"},
			{Name: "limit", Type: "int", Default: "25", Description: "page size"},
			{Name: "start", Type: "string", Description: "inclusive RFC3339 start"},
			{Name: "end", Type: "string", Description: "exclusive RFC3339 end"},
			{Name: "no-input", Type: "bool", Description: "never open a browser or prompt for input"},
			{Name: "safety-profile", Type: "string", Default: "readonly", Description: "automation policy (readonly only)"},
			{Name: "allow-command", Type: "string[]", Description: "allow only this exact command; repeatable"},
		},
		Commands: []commandSpec{
			{Name: "auth", Description: "authorize the WHOOP account", ReadOnly: true},
			{Name: "config diagnose", Description: "report credential configuration without secrets", ReadOnly: true},
			{Name: "profile", Description: "read the basic profile", ReadOnly: true},
			{Name: "body", Description: "read body measurements", ReadOnly: true},
			{Name: "cycles", Description: "read cycles", ReadOnly: true},
			{Name: "recovery", Description: "read recovery records", ReadOnly: true},
			{Name: "sleep", Description: "read sleep records", ReadOnly: true},
			{Name: "workouts", Description: "read workout records", ReadOnly: true},
			{Name: "brief", Description: "summarize latest recovery, sleep, and adherence", ReadOnly: true},
			{Name: "week", Description: "summarize workouts from the last seven days", ReadOnly: true},
			{Name: "weekly", Description: "summarize seven-day recovery, sleep, and adherence", ReadOnly: true},
			{Name: "schema", Description: "emit this machine-readable command contract", ReadOnly: true},
		},
	}
}

func runSchema(opts options) error {
	if opts.mode == output.Plain {
		return output.PlainValue(os.Stdout, commandSchema())
	}
	return output.JSONValue(os.Stdout, commandSchema())
}
