package whoop

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type APIError struct {
	StatusCode int
	Code       string
	Message    string
	RequestID  string
	RetryAfter time.Duration
}

func (e *APIError) Error() string {
	parts := []string{fmt.Sprintf("WHOOP API returned HTTP %d", e.StatusCode)}
	if e.Code != "" {
		parts = append(parts, "code="+Redact(e.Code))
	}
	if e.Message != "" {
		parts = append(parts, Redact(e.Message))
	}
	if e.RequestID != "" {
		parts = append(parts, "request_id="+Redact(e.RequestID))
	}
	return strings.Join(parts, ": ")
}
func (e *APIError) Retryable() bool {
	return e.StatusCode == http.StatusTooManyRequests || e.StatusCode >= 500
}

func parseAPIError(status int, headers http.Header, body []byte) *APIError {
	e := &APIError{StatusCode: status, RequestID: headers.Get("x-request-id"), RetryAfter: parseRetryAfter(headers.Get("retry-after"))}
	var payload struct {
		Error   string `json:"error"`
		Code    string `json:"code"`
		Message string `json:"message"`
		Detail  string `json:"detail"`
	}
	if json.Unmarshal(body, &payload) == nil {
		e.Code = payload.Code
		e.Message = payload.Message
		if e.Message == "" {
			e.Message = payload.Detail
		}
		if e.Message == "" {
			e.Message = payload.Error
		}
	}
	if e.Message == "" {
		e.Message = strings.TrimSpace(string(body))
	}
	e.Message = Redact(e.Message)
	if len(e.Message) > 300 {
		e.Message = e.Message[:300]
	}
	return e
}

func parseRetryAfter(value string) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(value); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	if when, err := http.ParseTime(value); err == nil {
		if d := time.Until(when); d > 0 {
			return d
		}
	}
	return 0
}
