package infakt

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var (
	// ErrNotFound is returned when a resource is not found (HTTP 404).
	ErrNotFound = errors.New("infakt: resource not found")

	// ErrUnauthorized is returned when the API key is invalid (HTTP 401).
	ErrUnauthorized = errors.New("infakt: unauthorized, check your API key")

	// ErrForbidden is returned when access is denied (HTTP 403).
	ErrForbidden = errors.New("infakt: forbidden")

	// ErrRateLimited is returned when the API rate limit is exceeded (HTTP 429).
	ErrRateLimited = errors.New("infakt: rate limit exceeded")
)

// ErrorResponse represents an error response from the inFakt API.
type ErrorResponse struct {
	Response   *http.Response `json:"-"`
	StatusCode int            `json:"status_code"`
	Message    string         `json:"message"`
	Method     string         `json:"-"`
	Endpoint   string         `json:"-"`
}

// Error implements the error interface.
func (e *ErrorResponse) Error() string {
	if e.Method != "" && e.Endpoint != "" {
		return fmt.Sprintf("infakt: %s %s: API error (status %d): %s",
			e.Method, e.Endpoint, e.StatusCode, e.Message)
	}
	return fmt.Sprintf("infakt: API error (status %d): %s", e.StatusCode, e.Message)
}

// Is allows using errors.Is with sentinel errors.
func (e *ErrorResponse) Is(target error) bool {
	switch e.StatusCode {
	case http.StatusNotFound:
		return target == ErrNotFound
	case http.StatusUnauthorized:
		return target == ErrUnauthorized
	case http.StatusForbidden:
		return target == ErrForbidden
	case http.StatusTooManyRequests:
		return target == ErrRateLimited
	}
	return false
}

// checkResponse checks the API response for errors and includes request
// context (method and endpoint) in the error for easier debugging.
func checkResponse(r *http.Response) error {
	if r.StatusCode >= 200 && r.StatusCode <= 299 {
		return nil
	}

	errResp := &ErrorResponse{
		Response:   r,
		StatusCode: r.StatusCode,
	}

	if r.Request != nil {
		errResp.Method = r.Request.Method
		errResp.Endpoint = r.Request.URL.Path
	}

	data, err := io.ReadAll(r.Body)
	if err == nil && len(data) > 0 {
		var apiErr struct {
			Message string `json:"message"`
			Error   string `json:"error"`
		}
		if json.Unmarshal(data, &apiErr) == nil {
			if apiErr.Message != "" {
				errResp.Message = apiErr.Message
			} else if apiErr.Error != "" {
				errResp.Message = apiErr.Error
			}
		}
	}

	if errResp.Message == "" {
		errResp.Message = http.StatusText(r.StatusCode)
	}

	return errResp
}
