// Package whoop provides a read-only client for the public WHOOP API.
package whoop

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultBaseURL   = "https://api.prod.whoop.com/developer"
	DefaultUserAgent = "whoop-cli/1.0"
)

type ClientConfig struct {
	BaseURL      string
	HTTPClient   *http.Client
	Timeout      time.Duration
	MaxRetries   int
	RetryBackoff func(attempt int) time.Duration
	Sleep        func(context.Context, time.Duration) error
	UserAgent    string
	AccessToken  string
}

type Client struct {
	baseURL      *url.URL
	httpClient   *http.Client
	timeout      time.Duration
	maxRetries   int
	retryBackoff func(int) time.Duration
	sleep        func(context.Context, time.Duration) error
	userAgent    string
	accessToken  string
}

func NewClient(cfg ClientConfig) (*Client, error) {
	base := cfg.BaseURL
	if base == "" {
		base = DefaultBaseURL
	}
	u, err := url.Parse(strings.TrimRight(base, "/"))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("invalid WHOOP base URL: %q", base)
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{}
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 0
	}
	if cfg.RetryBackoff == nil {
		cfg.RetryBackoff = func(attempt int) time.Duration {
			if attempt > 3 {
				attempt = 3
			}
			return time.Second << attempt
		}
	}
	if cfg.Sleep == nil {
		cfg.Sleep = func(ctx context.Context, d time.Duration) error {
			timer := time.NewTimer(d)
			defer timer.Stop()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				return nil
			}
		}
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = DefaultUserAgent
	}
	return &Client{baseURL: u, httpClient: cfg.HTTPClient, timeout: cfg.Timeout, maxRetries: cfg.MaxRetries, retryBackoff: cfg.RetryBackoff, sleep: cfg.Sleep, userAgent: cfg.UserAgent, accessToken: cfg.AccessToken}, nil
}

type ListOptions struct {
	Limit     int
	Start     string
	End       string
	NextToken string
}

func (o ListOptions) values(next string) url.Values {
	v := url.Values{}
	if o.Limit > 0 {
		v.Set("limit", strconv.Itoa(o.Limit))
	}
	if o.Start != "" {
		v.Set("start", o.Start)
	}
	if o.End != "" {
		v.Set("end", o.End)
	}
	if next != "" {
		v.Set("next_token", next)
	}
	return v
}

type Page[T any] struct {
	Records   []T    `json:"records"`
	NextToken string `json:"next_token"`
}

type PaginationError struct{ Token string }

func (e *PaginationError) Error() string { return "WHOOP returned a repeated pagination token" }

func listAll[T any](ctx context.Context, c *Client, path string, opts ListOptions) ([]T, error) {
	var all []T
	seen := map[string]struct{}{}
	next := opts.NextToken
	for {
		var page Page[T]
		if err := c.get(ctx, path, opts.values(next), &page); err != nil {
			return nil, err
		}
		all = append(all, page.Records...)
		if page.NextToken == "" {
			return all, nil
		}
		if _, ok := seen[page.NextToken]; ok {
			return nil, &PaginationError{Token: page.NextToken}
		}
		seen[page.NextToken] = struct{}{}
		next = page.NextToken
	}
}

func page[T any](ctx context.Context, c *Client, path string, opts ListOptions) (Page[T], error) {
	var result Page[T]
	err := c.get(ctx, path, opts.values(opts.NextToken), &result)
	return result, err
}

func (c *Client) BodyMeasurementsPage(ctx context.Context, opts ListOptions) (Page[BodyMeasurement], error) {
	return page[BodyMeasurement](ctx, c, "/v2/user/measurement/body", opts)
}

func (c *Client) CyclesPage(ctx context.Context, opts ListOptions) (Page[Cycle], error) {
	return page[Cycle](ctx, c, "/v2/cycle", opts)
}

func (c *Client) RecoveryPage(ctx context.Context, opts ListOptions) (Page[Recovery], error) {
	return page[Recovery](ctx, c, "/v2/recovery", opts)
}

func (c *Client) SleepPage(ctx context.Context, opts ListOptions) (Page[Sleep], error) {
	return page[Sleep](ctx, c, "/v2/activity/sleep", opts)
}

func (c *Client) WorkoutsPage(ctx context.Context, opts ListOptions) (Page[Workout], error) {
	return page[Workout](ctx, c, "/v2/activity/workout", opts)
}

func (c *Client) Profile(ctx context.Context) (Profile, error) {
	var out Profile
	err := c.get(ctx, "/v2/user/profile/basic", nil, &out)
	return out, err
}
func (c *Client) BodyMeasurements(ctx context.Context, opts ListOptions) ([]BodyMeasurement, error) {
	return listAll[BodyMeasurement](ctx, c, "/v2/user/measurement/body", opts)
}
func (c *Client) Cycles(ctx context.Context, opts ListOptions) ([]Cycle, error) {
	return listAll[Cycle](ctx, c, "/v2/cycle", opts)
}
func (c *Client) Recovery(ctx context.Context, opts ListOptions) ([]Recovery, error) {
	return listAll[Recovery](ctx, c, "/v2/recovery", opts)
}
func (c *Client) Sleep(ctx context.Context, opts ListOptions) ([]Sleep, error) {
	return listAll[Sleep](ctx, c, "/v2/activity/sleep", opts)
}
func (c *Client) Workouts(ctx context.Context, opts ListOptions) ([]Workout, error) {
	return listAll[Workout](ctx, c, "/v2/activity/workout", opts)
}

func (c *Client) get(ctx context.Context, path string, query url.Values, out any) error {
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("WHOOP path must be absolute: %q", path)
	}
	u := *c.baseURL
	u.Path = strings.TrimRight(c.baseURL.Path, "/") + path
	u.RawQuery = query.Encode()
	for attempt := 0; ; attempt++ {
		reqCtx, cancel := context.WithTimeout(ctx, c.timeout)
		req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, u.String(), nil)
		if err != nil {
			cancel()
			return err
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", c.userAgent)
		if c.accessToken != "" {
			req.Header.Set("Authorization", "Bearer "+c.accessToken)
		}
		resp, err := c.httpClient.Do(req)
		if err != nil {
			cancel()
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if attempt < c.maxRetries {
				if err := c.sleep(ctx, c.retryBackoff(attempt)); err != nil {
					return err
				}
				continue
			}
			return fmt.Errorf("WHOOP request failed: %w", err)
		}
		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		cancel()
		if readErr != nil {
			return fmt.Errorf("read WHOOP response: %w", readErr)
		}
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if out == nil || len(body) == 0 {
				return nil
			}
			if err := json.Unmarshal(body, out); err != nil {
				return fmt.Errorf("decode WHOOP response: %w", err)
			}
			return nil
		}
		apiErr := parseAPIError(resp.StatusCode, resp.Header, body)
		if apiErr.Retryable() && attempt < c.maxRetries {
			delay := apiErr.RetryAfter
			if delay <= 0 {
				delay = c.retryBackoff(attempt)
			}
			if err := c.sleep(ctx, delay); err != nil {
				return err
			}
			continue
		}
		return apiErr
	}
}

// Profile and resource values are intentionally strings for timestamp fields: the
// API occasionally returns null and its precision varies between endpoints.
type Profile struct {
	UserID    string `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Timezone  string `json:"timezone"`
}
type BodyMeasurement struct {
	ID             string  `json:"id"`
	CreatedAt      string  `json:"created_at"`
	HeightMeter    float64 `json:"height_meter"`
	WeightKilogram float64 `json:"weight_kilogram"`
	MaxHeartRate   int     `json:"max_heart_rate"`
}
type Cycle struct {
	ID     int64       `json:"id"`
	Start  string      `json:"start"`
	End    string      `json:"end"`
	Score  *CycleScore `json:"score"`
	UserID string      `json:"user_id"`
	Strain float64     `json:"strain"`
}
type CycleScore struct {
	Strain           float64 `json:"strain"`
	Kilojoule        float64 `json:"kilojoule"`
	AverageHeartRate int     `json:"average_heart_rate"`
	MaxHeartRate     int     `json:"max_heart_rate"`
}
type Recovery struct {
	CycleID    int64          `json:"cycle_id"`
	SleepID    string         `json:"sleep_id"`
	UserID     string         `json:"user_id"`
	CreatedAt  string         `json:"created_at"`
	Score      *RecoveryScore `json:"score"`
	ScoreState string         `json:"score_state"`
}
type RecoveryScore struct {
	RecoveryScore    float64 `json:"recovery_score"`
	HRVRmssdMilli    float64 `json:"hrv_rmssd_milli"`
	RestingHeartRate float64 `json:"resting_heart_rate"`
	SkinTempCelsius  float64 `json:"skin_temp_celsius"`
}
type Sleep struct {
	ID         string      `json:"id"`
	Start      string      `json:"start"`
	End        string      `json:"end"`
	Nap        bool        `json:"nap"`
	Score      *SleepScore `json:"score"`
	ScoreState string      `json:"score_state"`
}
type SleepScore struct {
	StageSummary               *StageSummary `json:"stage_summary"`
	SleepPerformancePercentage float64       `json:"sleep_performance_percentage"`
	SleepEfficiencyPercentage  float64       `json:"sleep_efficiency_percentage"`
	RespiratoryRate            float64       `json:"respiratory_rate"`
}
type StageSummary struct {
	TotalInBedTimeMilli         int64 `json:"total_in_bed_time_milli"`
	TotalLightSleepTimeMilli    int64 `json:"total_light_sleep_time_milli"`
	TotalSlowWaveSleepTimeMilli int64 `json:"total_slow_wave_sleep_time_milli"`
	TotalRemSleepTimeMilli      int64 `json:"total_rem_sleep_time_milli"`
	TotalAwakeTimeMilli         int64 `json:"total_awake_time_milli"`
}
type Workout struct {
	ID         string        `json:"id"`
	Start      string        `json:"start"`
	End        string        `json:"end"`
	SportName  string        `json:"sport_name"`
	SportID    int           `json:"sport_id"`
	Score      *WorkoutScore `json:"score"`
	ScoreState string        `json:"score_state"`
}
type WorkoutScore struct {
	Strain           float64 `json:"strain"`
	AverageHeartRate int     `json:"average_heart_rate"`
	MaxHeartRate     int     `json:"max_heart_rate"`
	Kilojoule        float64 `json:"kilojoule"`
}
