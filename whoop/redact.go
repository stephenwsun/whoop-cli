package whoop

import "regexp"

var (
	bearerPattern = regexp.MustCompile(`(?i)(bearer\s+)[A-Za-z0-9._~+\-/]+=*`)
	secretPattern = regexp.MustCompile(`(?i)((?:client[_-]?secret|refresh[_-]?token|access[_-]?token|authorization)["'\s:=]+)[^\s,;&]+`)
)

// Redact removes credentials from errors and diagnostics before they reach a terminal.
func Redact(value string) string {
	value = bearerPattern.ReplaceAllString(value, `${1}[REDACTED]`)
	return secretPattern.ReplaceAllString(value, `${1}[REDACTED]`)
}
