package infakt

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestErrorResponse_Error(t *testing.T) {
	e := &ErrorResponse{StatusCode: 404, Message: "Not Found"}
	expected := "infakt: API error (status 404): Not Found"
	if e.Error() != expected {
		t.Errorf("expected %q, got %q", expected, e.Error())
	}
}

func TestErrorResponse_Is(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		target     error
		want       bool
	}{
		{"400 is ErrBadRequest", 400, ErrBadRequest, true},
		{"401 is ErrUnauthorized", 401, ErrUnauthorized, true},
		{"402 is ErrPaymentRequired", 402, ErrPaymentRequired, true},
		{"403 is ErrForbidden", 403, ErrForbidden, true},
		{"404 is ErrNotFound", 404, ErrNotFound, true},
		{"422 is ErrUnprocessableEntity", 422, ErrUnprocessableEntity, true},
		{"423 is ErrLocked", 423, ErrLocked, true},
		{"429 is ErrRateLimited", 429, ErrRateLimited, true},
		{"503 is ErrRateLimited", 503, ErrRateLimited, true},
		{"404 is not ErrUnauthorized", 404, ErrUnauthorized, false},
		{"500 is not ErrNotFound", 500, ErrNotFound, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ErrorResponse{StatusCode: tt.statusCode, Message: "test"}
			if got := errors.Is(e, tt.target); got != tt.want {
				t.Errorf("errors.Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorResponse_As(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"invoice not found"}`))
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	req, _ := c.newRequest(context.Background(), http.MethodGet, "/test", nil)
	err := c.do(req, nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var errResp *ErrorResponse
	if !errors.As(err, &errResp) {
		t.Fatal("expected error to be *ErrorResponse")
	}

	if errResp.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", errResp.StatusCode)
	}
	if errResp.Message != "invoice not found" {
		t.Errorf("expected message %q, got %q", "invoice not found", errResp.Message)
	}
}

func TestErrorResponse_FieldErrors(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		// Mirrors a real inFakt 422: a top-level message plus a per-field
		// breakdown, with one field carrying a bare string instead of an array.
		_, _ = w.Write([]byte(`{"error":"Nieprawidłowe parametry.","errors":{"bank_name":["Proszę podać nazwę banku."],"base":"Ogólny błąd"}}`))
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	req, _ := c.newRequest(context.Background(), http.MethodPost, "/test", nil)
	err := c.do(req, nil)

	if !errors.Is(err, ErrUnprocessableEntity) {
		t.Fatalf("expected ErrUnprocessableEntity, got %v", err)
	}

	var errResp *ErrorResponse
	if !errors.As(err, &errResp) {
		t.Fatal("expected error to be *ErrorResponse")
	}
	if got := errResp.Errors["bank_name"]; len(got) != 1 || got[0] != "Proszę podać nazwę banku." {
		t.Errorf("bank_name errors = %v, want [\"Proszę podać nazwę banku.\"]", got)
	}
	if got := errResp.Errors["base"]; len(got) != 1 || got[0] != "Ogólny błąd" {
		t.Errorf("base errors (bare string coercion) = %v, want [\"Ogólny błąd\"]", got)
	}
	// The field detail must be visible in the error string.
	if !strings.Contains(err.Error(), "bank_name: Proszę podać nazwę banku.") {
		t.Errorf("Error() = %q, want it to contain the field detail", err.Error())
	}
}

func TestErrorResponse_FieldErrorsUnexpectedShape(t *testing.T) {
	// If "errors" arrives as an array (not the documented object), the
	// top-level message must still be extracted and Errors left nil.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"error":"Coś poszło nie tak","errors":["luźny komunikat"]}`))
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	req, _ := c.newRequest(context.Background(), http.MethodPost, "/test", nil)
	err := c.do(req, nil)

	var errResp *ErrorResponse
	if !errors.As(err, &errResp) {
		t.Fatal("expected error to be *ErrorResponse")
	}
	if errResp.Message != "Coś poszło nie tak" {
		t.Errorf("message = %q, want %q (must survive unexpected errors shape)", errResp.Message, "Coś poszło nie tak")
	}
	if errResp.Errors != nil {
		t.Errorf("Errors = %v, want nil for non-object shape", errResp.Errors)
	}
}

func TestCheckResponse_Success(t *testing.T) {
	resp := &http.Response{StatusCode: 200}
	if err := checkResponse(resp); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestCheckResponse_ErrorWithJSONBody(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid api key"}`))
	}))
	defer ts.Close()

	c := newTestClient("bad-key", WithBaseURL(ts.URL))
	req, _ := c.newRequest(context.Background(), http.MethodGet, "/test", nil)
	err := c.do(req, nil)

	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}
