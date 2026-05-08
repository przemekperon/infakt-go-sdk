package infakt

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// The following sentinel errors form the family of HTTP-status-derived
// errors returned by this package. They are never returned directly by
// API service methods; instead, methods return an [*ErrorResponse] whose
// [ErrorResponse.Is] method maps an HTTP status code to the appropriate
// sentinel. Use [errors.Is] to check the kind of failure, e.g.
//
//	if errors.Is(err, infakt.ErrNotFound) { ... }
var (
	// ErrBadRequest is returned when the request is malformed (HTTP 400).
	ErrBadRequest = errors.New("infakt: bad request")

	// ErrUnauthorized is returned when the API key is invalid (HTTP 401).
	ErrUnauthorized = errors.New("infakt: unauthorized, check your API key")

	// ErrPaymentRequired is returned when the account plan limit has been
	// reached and an upgrade is required (HTTP 402).
	ErrPaymentRequired = errors.New("infakt: payment required (plan limit reached)")

	// ErrForbidden is returned when access is denied (HTTP 403).
	ErrForbidden = errors.New("infakt: forbidden")

	// ErrNotFound is returned when a resource is not found (HTTP 404).
	ErrNotFound = errors.New("infakt: resource not found")

	// ErrUnprocessableEntity is returned when the request payload fails
	// server-side validation (HTTP 422).
	ErrUnprocessableEntity = errors.New("infakt: unprocessable entity")

	// ErrLocked is returned when the resource is temporarily locked
	// (HTTP 423).
	ErrLocked = errors.New("infakt: resource locked")

	// ErrRateLimited is returned when the API rate limit is exceeded
	// (HTTP 429) or the IP-level limit triggers HTTP 503.
	ErrRateLimited = errors.New("infakt: rate limit exceeded")
)

// ErrorResponse represents an error response from the inFakt API.
type ErrorResponse struct {
	// Response is the underlying HTTP response that produced this error.
	// It may be nil if the error was constructed without a response (for
	// example, from synthetic test fixtures).
	Response *http.Response `json:"-"`
	// StatusCode is the HTTP status code that triggered the error.
	StatusCode int `json:"status_code"`
	// Message is the server-provided error message, if any; otherwise a
	// synthesized description derived from the HTTP status text. The
	// inFakt API returns the message in a JSON field named "error".
	Message string `json:"error"`
	// Method is the HTTP method (GET, POST, ...) of the failing request.
	Method string `json:"-"`
	// Endpoint is the request path of the failing request, included in
	// the error string for easier debugging.
	Endpoint string `json:"-"`
}

// Error implements the error interface.
func (e *ErrorResponse) Error() string {
	if e.Method != "" && e.Endpoint != "" {
		return fmt.Sprintf("infakt: %s %s: API error (status %d): %s",
			e.Method, e.Endpoint, e.StatusCode, e.Message)
	}
	return fmt.Sprintf("infakt: API error (status %d): %s", e.StatusCode, e.Message)
}

// Is reports whether this *ErrorResponse should be considered equivalent to
// one of the package sentinel errors when compared with [errors.Is]. It maps
// the HTTP [ErrorResponse.StatusCode] to the matching sentinel:
//
//   - 400 -> [ErrBadRequest]
//   - 401 -> [ErrUnauthorized]
//   - 402 -> [ErrPaymentRequired]
//   - 403 -> [ErrForbidden]
//   - 404 -> [ErrNotFound]
//   - 422 -> [ErrUnprocessableEntity]
//   - 423 -> [ErrLocked]
//   - 429, 503 -> [ErrRateLimited]
func (e *ErrorResponse) Is(target error) bool {
	switch e.StatusCode {
	case http.StatusBadRequest:
		return target == ErrBadRequest
	case http.StatusUnauthorized:
		return target == ErrUnauthorized
	case http.StatusPaymentRequired:
		return target == ErrPaymentRequired
	case http.StatusForbidden:
		return target == ErrForbidden
	case http.StatusNotFound:
		return target == ErrNotFound
	case http.StatusUnprocessableEntity:
		return target == ErrUnprocessableEntity
	case http.StatusLocked:
		return target == ErrLocked
	case http.StatusTooManyRequests, http.StatusServiceUnavailable:
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
		// inFakt returns errors as {"error":"..."} per https://docs.infakt.pl
		// (sekcja "Kody błędów"). Older snapshots and some 5xx fronting
		// proxies use {"message":"..."}; accept both for forward and
		// backward compatibility.
		var apiErr struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		if json.Unmarshal(data, &apiErr) == nil {
			switch {
			case apiErr.Error != "":
				errResp.Message = apiErr.Error
			case apiErr.Message != "":
				errResp.Message = apiErr.Message
			}
		}
	}

	if errResp.Message == "" {
		errResp.Message = http.StatusText(r.StatusCode)
	}

	return errResp
}
