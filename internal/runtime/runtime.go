package runtime

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/stephensun/whoop-cli/internal/auth"
	"github.com/stephensun/whoop-cli/whoop"
)

type Config struct {
	Store        auth.Store
	ClientID     string
	ClientSecret string
}

type Diagnostics struct {
	Store                  string `json:"store"`
	ClientIDConfigured     bool   `json:"client_id_configured"`
	ClientSecretConfigured bool   `json:"client_secret_configured"`
	RefreshTokenConfigured bool   `json:"refresh_token_configured"`
}

func Load(ctx context.Context) (Config, error) {
	store := auth.DefaultStore()
	clientID := strings.TrimSpace(os.Getenv("WHOOP_CLIENT_ID"))
	clientSecret := strings.TrimSpace(os.Getenv("WHOOP_CLIENT_SECRET"))
	var err error
	if clientID == "" {
		clientID, err = store.Get(ctx, "client_id")
		if err != nil && !errors.Is(err, auth.ErrNotFound) {
			return Config{}, err
		}
	}
	if clientSecret == "" {
		clientSecret, err = store.Get(ctx, "client_secret")
		if err != nil && !errors.Is(err, auth.ErrNotFound) {
			return Config{}, err
		}
	}
	if clientID == "" || clientSecret == "" {
		return Config{}, errors.New("WHOOP client credentials are missing; configure WHOOP_CLIENT_ID/WHOOP_CLIENT_SECRET or the secure credential store")
	}
	return Config{Store: store, ClientID: clientID, ClientSecret: clientSecret}, nil
}

func (c Config) OAuth() (*auth.OAuth, error) {
	return auth.New(auth.Config{ClientID: c.ClientID, ClientSecret: c.ClientSecret, Store: c.Store})
}

func (c Config) API(ctx context.Context) (*whoop.Client, error) {
	oauthClient, err := c.OAuth()
	if err != nil {
		return nil, err
	}
	token, err := oauthClient.Refresh(ctx)
	if err != nil {
		return nil, err
	}
	return whoop.NewClient(whoop.ClientConfig{HTTPClient: &http.Client{}, AccessToken: token.AccessToken, MaxRetries: 3})
}

func Diagnose(ctx context.Context) Diagnostics {
	store := auth.DefaultStore()
	_, idErr := store.Get(ctx, "client_id")
	_, secretErr := store.Get(ctx, "client_secret")
	_, refreshErr := store.Get(ctx, "refresh_token")
	return Diagnostics{
		Store:                  fmt.Sprintf("%T", store),
		ClientIDConfigured:     os.Getenv("WHOOP_CLIENT_ID") != "" || idErr == nil,
		ClientSecretConfigured: os.Getenv("WHOOP_CLIENT_SECRET") != "" || secretErr == nil,
		RefreshTokenConfigured: refreshErr == nil,
	}
}
