package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/stephensun/whoop-cli/internal/output"
	"github.com/stephensun/whoop-cli/internal/runtime"
	"github.com/stephensun/whoop-cli/internal/summary"
	"github.com/stephensun/whoop-cli/whoop"
)

type options struct {
	mode          output.Mode
	limit         int
	start, end    string
	noInput       bool
	safetyProfile string
	allowCommands []string
}

type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ",") }
func (s *stringList) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
func validatePolicy(command, configSubcommand string, opts options) error {
	if opts.safetyProfile != "readonly" {
		return fmt.Errorf("unsupported safety profile %q; only readonly is available", opts.safetyProfile)
	}
	policyCommand := command
	if configSubcommand != "" {
		policyCommand += " " + configSubcommand
	}
	if len(opts.allowCommands) > 0 && !contains(opts.allowCommands, policyCommand) {
		return fmt.Errorf("command %q is not allowed by --allow-command", policyCommand)
	}
	if opts.noInput && command == "auth" {
		return errors.New("auth requires browser input and cannot run with --no-input")
	}
	return nil
}
func Run(ctx context.Context, args []string) {
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h" || args[0] == "help") {
		usage()
		return
	}
	command := "brief"
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		command, args = args[0], args[1:]
	}
	if command == "help" || command == "--help" {
		usage()
		return
	}
	configSubcommand := ""
	if command == "config" {
		filtered := make([]string, 0, len(args))
		for _, arg := range args {
			if arg == "diagnose" {
				configSubcommand = arg
				continue
			}
			filtered = append(filtered, arg)
		}
		args = filtered
	}
	fs := flag.NewFlagSet("whoop "+command, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	jsonFlag := fs.Bool("json", false, "emit stable JSON")
	plainFlag := fs.Bool("plain", false, "emit stable tab-separated output")
	limit := fs.Int("limit", 25, "page size")
	start := fs.String("start", "", "inclusive RFC3339 start")
	end := fs.String("end", "", "exclusive RFC3339 end")
	noInput := fs.Bool("no-input", false, "never open a browser or prompt for input")
	safetyProfile := fs.String("safety-profile", "readonly", "automation policy (readonly only)")
	var allowCommands stringList
	fs.Var(&allowCommands, "allow-command", "allow only this exact command (repeatable)")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	if *jsonFlag && *plainFlag {
		fail(2, errors.New("--json and --plain are mutually exclusive"))
	}
	mode := output.Human
	if *jsonFlag {
		mode = output.JSON
	}
	if *plainFlag {
		mode = output.Plain
	}
	opts := options{mode: mode, limit: *limit, start: *start, end: *end, noInput: *noInput, safetyProfile: *safetyProfile, allowCommands: allowCommands}
	if err := validatePolicy(command, configSubcommand, opts); err != nil {
		fail(2, err)
	}
	var err error
	switch command {
	case "schema":
		err = runSchema(opts)
	case "auth":
		err = runAuth(ctx)
	case "config":
		err = runConfig(ctx, []string{configSubcommand}, opts)
	case "profile":
		err = runProfile(ctx, opts)
	case "body", "body-measurements", "measurements":
		err = runList(ctx, "body measurements", opts)
	case "cycles":
		err = runList(ctx, "cycles", opts)
	case "recovery":
		err = runList(ctx, "recovery", opts)
	case "sleep":
		err = runList(ctx, "sleep", opts)
	case "workouts":
		err = runList(ctx, "workouts", opts)
	case "brief":
		err = runBrief(ctx, opts)
	case "week":
		err = runWeek(ctx, opts)
	case "weekly":
		err = runWeekly(ctx, opts)
	default:
		usage()
		fail(2, fmt.Errorf("unknown command %q", command))
	}
	if err != nil {
		fail(1, err)
	}
}

func usage() {
	_, _ = fmt.Fprintln(os.Stderr, `whoop: read-only WHOOP account data CLI

Usage: whoop <command> [--json|--plain] [--limit N] [--start RFC3339] [--end RFC3339]

	Commands: auth, config diagnose, schema, profile, body, cycles, recovery, sleep, workouts, brief, week, weekly

Configuration: WHOOP_CLIENT_ID and WHOOP_CLIENT_SECRET may be environment variables;
otherwise they are read from the secure credential store under client_id/client_secret.`)
}
func fail(code int, err error) { output.Error(os.Stderr, err); os.Exit(code) }

func newAPI(ctx context.Context) (*whoop.Client, error) {
	config, err := runtime.Load(ctx)
	if err != nil {
		return nil, err
	}
	return config.API(ctx)
}

func runAuth(ctx context.Context) error {
	config, err := runtime.Load(ctx)
	if err != nil {
		return err
	}
	client, err := config.OAuth()
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "whoop: opening browser for WHOOP consent; approve access there...")
	if err := client.Authenticate(ctx); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "whoop: authorized; refresh token stored securely")
	return nil
}

func runConfig(ctx context.Context, args []string, opts options) error {
	if len(args) == 0 || args[0] != "diagnose" {
		return errors.New("usage: whoop config diagnose")
	}
	d := runtime.Diagnose(ctx)
	if opts.mode == output.JSON {
		return output.JSONValue(os.Stdout, d)
	}
	if opts.mode == output.Plain {
		return output.PlainValue(os.Stdout, d)
	}
	_, err := fmt.Fprintf(os.Stdout, "store | %s\nclient_id | configured=%t\nclient_secret | configured=%t\nrefresh_token | configured=%t\n", d.Store, d.ClientIDConfigured, d.ClientSecretConfigured, d.RefreshTokenConfigured)
	return err
}

func listOptions(opts options) whoop.ListOptions {
	return whoop.ListOptions{Limit: opts.limit, Start: opts.start, End: opts.end}
}
func runProfile(ctx context.Context, opts options) error {
	api, err := newAPI(ctx)
	if err != nil {
		return err
	}
	profile, err := api.Profile(ctx)
	if err != nil {
		return err
	}
	return emit(opts, profile, "profile")
}
func runList(ctx context.Context, name string, opts options) error {
	api, err := newAPI(ctx)
	if err != nil {
		return err
	}
	var value any
	switch name {
	case "body measurements":
		value, err = api.BodyMeasurements(ctx, listOptions(opts))
	case "cycles":
		value, err = api.Cycles(ctx, listOptions(opts))
	case "recovery":
		value, err = api.Recovery(ctx, listOptions(opts))
	case "sleep":
		value, err = api.Sleep(ctx, listOptions(opts))
	case "workouts":
		value, err = api.Workouts(ctx, listOptions(opts))
	}
	if err != nil {
		return err
	}
	return emit(opts, value, name)
}
func emit(opts options, value any, name string) error {
	switch opts.mode {
	case output.JSON:
		return output.JSONValue(os.Stdout, value)
	case output.Plain:
		return output.PlainValue(os.Stdout, value)
	default:
		return output.PlainValue(os.Stdout, value)
	}
}

func weekWindow(opts options) whoop.ListOptions {
	if opts.start == "" {
		opts.start = time.Now().UTC().Add(-7 * 24 * time.Hour).Format(time.RFC3339)
	}
	return listOptions(opts)
}
func runBrief(ctx context.Context, opts options) error {
	api, err := newAPI(ctx)
	if err != nil {
		return err
	}
	recPage, err := api.RecoveryPage(ctx, whoop.ListOptions{Limit: 1})
	if err != nil {
		return err
	}
	sleepPage, err := api.SleepPage(ctx, whoop.ListOptions{Limit: 1})
	if err != nil {
		return err
	}
	workouts, err := api.Workouts(ctx, weekWindow(opts))
	if err != nil {
		return err
	}
	return emitLines(opts, summary.Brief(recPage.Records, sleepPage.Records, workouts))
}
func runWeek(ctx context.Context, opts options) error {
	api, err := newAPI(ctx)
	if err != nil {
		return err
	}
	workouts, err := api.Workouts(ctx, weekWindow(opts))
	if err != nil {
		return err
	}
	if opts.mode == output.JSON {
		return output.JSONValue(os.Stdout, workouts)
	}
	if len(workouts) == 0 {
		_, err = fmt.Fprintln(os.Stdout, "no workouts recorded in the last 7 days")
		return err
	}
	for _, workout := range workouts {
		strain := ""
		if workout.Score != nil {
			strain = fmt.Sprintf(" | strain %.1f", workout.Score.Strain)
		}
		if _, err := fmt.Fprintf(os.Stdout, "%s | %s%s\n", date(workout.Start), workout.SportName, strain); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(os.Stdout, summary.Adherence(workouts))
	return err
}
func runWeekly(ctx context.Context, opts options) error {
	api, err := newAPI(ctx)
	if err != nil {
		return err
	}
	window := weekWindow(opts)
	rec, err := api.Recovery(ctx, window)
	if err != nil {
		return err
	}
	sleeps, err := api.Sleep(ctx, window)
	if err != nil {
		return err
	}
	workouts, err := api.Workouts(ctx, window)
	if err != nil {
		return err
	}
	return emitLines(opts, summary.Weekly(rec, sleeps, workouts))
}
func emitLines(opts options, lines []string) error {
	if opts.mode == output.JSON {
		return output.JSONValue(os.Stdout, struct {
			Lines []string `json:"lines"`
		}{lines})
	}
	for _, line := range lines {
		if _, err := fmt.Fprintln(os.Stdout, line); err != nil {
			return err
		}
	}
	return nil
}
func date(value string) string {
	if len(value) >= 10 {
		return value[:10]
	}
	return value
}
