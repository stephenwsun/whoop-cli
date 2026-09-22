// Package version contains build metadata exposed by the whoop binary.
package version

import "fmt"

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

type Info struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
}

func Current() Info {
	return Info{Version: Version, Commit: Commit, Date: Date}
}

func (i Info) String() string {
	return fmt.Sprintf("%s (commit %s, built %s)", i.Version, i.Commit, i.Date)
}

func UserAgent() string {
	return "whoop-cli/" + Version
}
