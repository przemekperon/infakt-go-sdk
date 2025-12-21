package infakt

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
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
		{"404 is ErrNotFound", 404, ErrNotFound, true},
		{"401 is ErrUnauthorized", 401, ErrUnauthorized, true},
		{"403 is ErrForbidden", 403, ErrForbidden, true},
		{"429 is ErrRateLimited", 429, ErrRateLimited, true},
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
		w.Write([]byte(`{"message":"invoice not found"}`))
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	req, _ := c.newRequest(context.Background(), http.MethodGet, "/test", nil)
	_, err := c.do(req, nil)

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
		w.Write([]byte(`{"error":"invalid api key"}`))
	}))
	defer ts.Close()

	c := newTestClient("bad-key", WithBaseURL(ts.URL))
	req, _ := c.newRequest(context.Background(), http.MethodGet, "/test", nil)
	_, err := c.do(req, nil)

	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}
