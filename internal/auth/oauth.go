package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/stephenwsun/whoop-cli/internal/version"
)

const (
	DefaultAuthURL     = "https://api.prod.whoop.com/oauth/oauth2/auth"
	DefaultTokenURL    = "https://api.prod.whoop.com/oauth/oauth2/token"
	DefaultRedirectURL = "http://localhost:8400/callback"
	DefaultScopes      = "read:profile read:body_measurement read:cycles read:recovery read:sleep read:workout offline"
)

type Config struct {
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	RedirectURL  string
	Scopes       string
	HTTPClient   *http.Client
	Store        Store
	OpenBrowser  func(string) error
}

type OAuth struct{ cfg Config }
type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}
type TokenError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *TokenError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("WHOOP OAuth endpoint returned HTTP %d", e.StatusCode)
	}
	return fmt.Sprintf("WHOOP OAuth endpoint returned HTTP %d: %s", e.StatusCode, redact(e.Message))
}

func New(cfg Config) (*OAuth, error) {
	if strings.TrimSpace(cfg.ClientID) == "" || strings.TrimSpace(cfg.ClientSecret) == "" {
		return nil, errors.New("WHOOP OAuth client credentials are not configured")
	}
	if cfg.AuthURL == "" {
		cfg.AuthURL = DefaultAuthURL
	}
	if cfg.TokenURL == "" {
		cfg.TokenURL = DefaultTokenURL
	}
	if cfg.RedirectURL == "" {
		cfg.RedirectURL = DefaultRedirectURL
	}
	if cfg.Scopes == "" {
		cfg.Scopes = DefaultScopes
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}
	if cfg.Store == nil {
		return nil, errors.New("WHOOP credential store is not configured")
	}
	if cfg.OpenBrowser == nil {
		cfg.OpenBrowser = openBrowser
	}
	return &OAuth{cfg: cfg}, nil
}

func (o *OAuth) AuthorizationURL(state string) (string, error) {
	u, err := url.Parse(o.cfg.AuthURL)
	if err != nil {
		return "", fmt.Errorf("invalid WHOOP auth URL: %w", err)
	}
	q := u.Query()
	q.Set("client_id", o.cfg.ClientID)
	q.Set("redirect_uri", o.cfg.RedirectURL)
	q.Set("response_type", "code")
	q.Set("scope", o.cfg.Scopes)
	q.Set("state", state)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (o *OAuth) ExchangeCode(ctx context.Context, code string) (Token, error) {
	return o.tokenRequest(ctx, url.Values{"grant_type": {"authorization_code"}, "code": {code}, "client_id": {o.cfg.ClientID}, "client_secret": {o.cfg.ClientSecret}, "redirect_uri": {o.cfg.RedirectURL}})
}
func (o *OAuth) Refresh(ctx context.Context) (Token, error) {
	refresh, err := o.cfg.Store.Get(ctx, "refresh_token")
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Token{}, errors.New("WHOOP is not authorized; run `whoop auth`")
		}
		return Token{}, err
	}
	token, err := o.tokenRequest(ctx, url.Values{"grant_type": {"refresh_token"}, "refresh_token": {refresh}, "client_id": {o.cfg.ClientID}, "client_secret": {o.cfg.ClientSecret}, "scope": {"offline"}})
	if err != nil {
		return Token{}, err
	}
	if token.RefreshToken == "" {
		return Token{}, errors.New("WHOOP refresh response did not include a rotated refresh token")
	}
	if err := o.cfg.Store.Set(ctx, "refresh_token", token.RefreshToken); err != nil {
		return Token{}, fmt.Errorf("store rotated WHOOP refresh token: %w", err)
	}
	return token, nil
}

func (o *OAuth) Authenticate(ctx context.Context) error {
	state, err := RandomState()
	if err != nil {
		return fmt.Errorf("generate OAuth state: %w", err)
	}
	u, err := o.AuthorizationURL(state)
	if err != nil {
		return err
	}
	redirect, err := url.Parse(o.cfg.RedirectURL)
	if err != nil {
		return fmt.Errorf("invalid redirect URL: %w", err)
	}
	listener, err := net.Listen("tcp", redirect.Host)
	if err != nil {
		return fmt.Errorf("listen for WHOOP OAuth callback: %w", err)
	}
	defer listener.Close()
	result := make(chan callbackResult, 1)
	server := &http.Server{
		ReadHeaderTimeout: 5 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != redirect.Path {
				http.NotFound(w, r)
				return
			}
			q := r.URL.Query()
			select {
			case result <- callbackResult{code: q.Get("code"), state: q.Get("state"), oauthError: q.Get("error")}:
			default:
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = io.WriteString(w, "<h2>WHOOP is connected. You can close this tab.</h2>")
		}),
	}
	go func() { _ = server.Serve(listener) }()
	if err := o.cfg.OpenBrowser(u); err != nil {
		_ = server.Shutdown(context.Background())
		return fmt.Errorf("open WHOOP authorization URL: %w", err)
	}
	var callback callbackResult
	select {
	case <-ctx.Done():
		_ = server.Shutdown(context.Background())
		return ctx.Err()
	case callback = <-result:
	}
	_ = server.Shutdown(context.Background())
	if callback.oauthError != "" {
		return fmt.Errorf("WHOOP authorization denied: %s", redact(callback.oauthError))
	}
	if callback.code == "" || callback.state != state {
		return errors.New("WHOOP OAuth callback state mismatch or missing code")
	}
	token, err := o.ExchangeCode(ctx, callback.code)
	if err != nil {
		return err
	}
	if token.RefreshToken == "" {
		return errors.New("WHOOP authorization response did not include a refresh token; enable offline access")
	}
	if err := o.cfg.Store.Set(ctx, "refresh_token", token.RefreshToken); err != nil {
		return fmt.Errorf("store WHOOP refresh token: %w", err)
	}
	return nil
}

type callbackResult struct{ code, state, oauthError string }

func openBrowser(rawURL string) error {
	if runtime.GOOS == "darwin" {
		return exec.Command("open", rawURL).Run()
	}
	if runtime.GOOS == "linux" {
		return exec.Command("xdg-open", rawURL).Run()
	}
	return fmt.Errorf("automatic browser opening is unsupported on %s; rerun `whoop auth --no-browser` and open the printed URL manually", runtime.GOOS)
}

var secretValuePattern = regexp.MustCompile(`(?i)((?:client[_-]?secret|refresh[_-]?token|access[_-]?token)\s*[=:]\s*)[^\s&,}"]+`)

func redact(s string) string {
	if len(s) > 200 {
		s = s[:200]
	}
	return secretValuePattern.ReplaceAllString(s, `${1}[REDACTED]`)
}

func (o *OAuth) tokenRequest(ctx context.Context, form url.Values) (Token, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", version.UserAgent())
	resp, err := o.cfg.HTTPClient.Do(req)
	if err != nil {
		return Token{}, fmt.Errorf("WHOOP OAuth request failed: %w", err)
	}
	defer resp.Body.Close()
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return Token{}, fmt.Errorf("read WHOOP OAuth response: %w", readErr)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var p struct {
			Error       string `json:"error"`
			Description string `json:"error_description"`
			Message     string `json:"message"`
		}
		_ = json.Unmarshal(body, &p)
		msg := p.Description
		if msg == "" {
			msg = p.Message
		}
		if msg == "" {
			msg = p.Error
		}
		return Token{}, &TokenError{StatusCode: resp.StatusCode, Code: redact(p.Error), Message: redact(msg)}
	}
	var token Token
	if err := json.Unmarshal(body, &token); err != nil {
		return Token{}, fmt.Errorf("decode WHOOP OAuth response: %w", err)
	}
	if token.AccessToken == "" {
		return Token{}, errors.New("WHOOP OAuth response did not include an access token")
	}
	return token, nil
}
