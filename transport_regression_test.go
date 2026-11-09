package infakt

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type regressionTransport func(*http.Request) (*http.Response, error)

func (f regressionTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func regressionResponse(req *http.Request, status int, body io.ReadCloser) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: body, Request: req}
}

func TestWritesAreNeverRetried(t *testing.T) {
	for _, status := range []int{500, 503, 429} {
		for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch} {
			t.Run(method+http.StatusText(status), func(t *testing.T) {
				calls := 0
				c := newTestClient("fake-key", WithHTTPClient(&http.Client{Transport: regressionTransport(func(r *http.Request) (*http.Response, error) {
					calls++
					return regressionResponse(r, status, io.NopCloser(strings.NewReader(`{"error":"ambiguous","errors":{"test":["detail"]}}`))), nil
				})}))
				defer c.Close()
				req, _ := c.newRequest(context.Background(), method, "/test", map[string]string{"test": "value"})
				err := c.do(req, nil)
				var apiErr *ErrorResponse
				if !errors.As(err, &apiErr) || apiErr.StatusCode != status || apiErr.Message != "ambiguous" || len(apiErr.Errors["test"]) != 1 {
					t.Fatalf("expected original API error: %v", err)
				}
				if status == 429 || status == 503 {
					if !errors.Is(err, ErrRateLimited) {
						t.Fatalf("sentinel lost: %v", err)
					}
				}
				if calls != 1 {
					t.Fatalf("write attempted %d times", calls)
				}
			})
		}
	}
}

func TestInvoiceSideEffectsAttemptedOnce(t *testing.T) {
	for _, action := range []string{"create", "email", "pdf"} {
		t.Run(action, func(t *testing.T) {
			var saved int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost && action != "pdf" {
					t.Errorf("unexpected method %s", r.Method)
				}
				atomic.AddInt32(&saved, 1) // The operation takes effect before the server fails.
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"response lost after saving"}`))
			}))
			defer server.Close()
			c := newTestClient("fake-key", WithBaseURL(server.URL))
			defer c.Close()
			var err error
			switch action {
			case "create":
				_, err = c.Invoices.Create(context.Background(), &Invoice{})
			case "pdf":
				_, err = c.Invoices.GetPDF(context.Background(), "test", "", "")
			default:
				err = c.Invoices.SendByEmail(context.Background(), "test", nil)
			}
			if err == nil || atomic.LoadInt32(&saved) != 1 {
				t.Fatalf("err=%v, side effects=%d", err, saved)
			}
		})
	}
}

func TestReadRetryBounds(t *testing.T) {
	for _, status := range []int{500, 429} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			calls := 0
			c := newTestClient("fake-key", WithHTTPClient(&http.Client{Transport: regressionTransport(func(r *http.Request) (*http.Response, error) {
				calls++
				resp := regressionResponse(r, status, io.NopCloser(strings.NewReader(`{"error":"still failing"}`)))
				resp.Header.Set("Retry-After", "0")
				return resp, nil
			})}))
			defer c.Close()
			req, _ := c.newRequest(context.Background(), http.MethodHead, "/test", nil)
			err := c.do(req, nil)
			expected := 3
			if status == 429 {
				expected = 2
			}
			if err == nil || calls != expected {
				t.Fatalf("err=%v attempts=%d want %d", err, calls, expected)
			}
		})
	}
}

type regressionBody struct {
	io.Reader
	closed bool
}

func (b *regressionBody) Close() error { b.closed = true; return nil }

func TestRetryReplaysBody(t *testing.T) {
	var payloads []string
	var bodies []*regressionBody
	c := newTestClient("fake-key", WithHTTPClient(&http.Client{Transport: regressionTransport(func(r *http.Request) (*http.Response, error) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		_ = r.Body.Close()
		payloads = append(payloads, string(data))
		if r.ContentLength != int64(len(data)) || r.Header.Get("Custom") != "kept" {
			t.Error("request metadata lost")
		}
		status := http.StatusOK
		if len(payloads) == 1 {
			status = http.StatusTooManyRequests
		}
		body := &regressionBody{Reader: strings.NewReader(`{}`)}
		bodies = append(bodies, body)
		resp := regressionResponse(r, status, body)
		resp.Header.Set("Retry-After", "0")
		return resp, nil
	})}))
	defer c.Close()
	req, _ := c.newRequest(context.Background(), http.MethodGet, "/test", map[string]string{"note": "exact"})
	req.Header.Set("Custom", "kept")
	if err := c.do(req, nil); err != nil {
		t.Fatal(err)
	}
	if len(payloads) != 2 || payloads[0] != payloads[1] || payloads[0] == "" {
		t.Fatalf("payloads: %q", payloads)
	}
	for _, body := range bodies {
		if !body.closed {
			t.Error("response body left open")
		}
	}
}

func TestRetryBodyUnavailable(t *testing.T) {
	for _, mode := range []string{"missing", "error"} {
		t.Run(mode, func(t *testing.T) {
			calls := 0
			body := &regressionBody{Reader: strings.NewReader(`{}`)}
			replayErr := errors.New("cannot recreate body")
			c := newTestClient("fake-key", WithHTTPClient(&http.Client{Transport: regressionTransport(func(r *http.Request) (*http.Response, error) {
				calls++
				_, _ = io.Copy(io.Discard, r.Body)
				_ = r.Body.Close()
				resp := regressionResponse(r, http.StatusTooManyRequests, body)
				resp.Header.Set("Retry-After", "0")
				return resp, nil
			})}))
			defer c.Close()
			req, _ := c.newRequest(context.Background(), http.MethodGet, "/test", map[string]string{"x": "y"})
			if mode == "missing" {
				req.GetBody = nil
			} else {
				req.GetBody = func() (io.ReadCloser, error) { return nil, replayErr }
			}
			err := c.do(req, nil)
			if mode == "missing" && !errors.Is(err, ErrRateLimited) {
				t.Fatalf("original API error lost: %v", err)
			}
			if mode == "error" && !errors.Is(err, replayErr) {
				t.Fatalf("replay error lost: %v", err)
			}
			if calls != 1 || !body.closed {
				t.Fatalf("calls=%d closed=%v", calls, body.closed)
			}
		})
	}
}

func TestCancellationDuringBackoff(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	received := make(chan struct{})
	var calls int32
	c := newTestClient("fake-key", WithHTTPClient(&http.Client{Transport: regressionTransport(func(r *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)
		close(received)
		resp := regressionResponse(r, 429, io.NopCloser(strings.NewReader(`{}`)))
		resp.Header.Set("Retry-After", "60")
		return resp, nil
	})}))
	defer c.Close()
	req, _ := c.newRequest(ctx, http.MethodGet, "/test", nil)
	done := make(chan error, 1)
	go func() { done <- c.do(req, nil) }()
	<-received
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("backoff ignored cancellation")
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("attempts=%d", calls)
	}
}

func TestCancellationBeforeAndDuringLimiter(t *testing.T) {
	for _, raw := range []bool{false, true} {
		for _, limited := range []bool{false, true} {
			for _, preCancelled := range []bool{false, true} {
				t.Run(strings.Join([]string{map[bool]string{true: "raw", false: "json"}[raw], map[bool]string{true: "limited", false: "unlimited"}[limited], map[bool]string{true: "before", false: "during"}[preCancelled]}, "/"), func(t *testing.T) {
					// Unlimited calls are tested only with cancellation before dispatch.
					if !limited && !preCancelled {
						return
					}
					var calls int32
					interval := time.Duration(0)
					if limited {
						interval = time.Hour
					}
					c := newTestClient("fake-key", WithRateLimit(interval), WithHTTPClient(&http.Client{Transport: regressionTransport(func(r *http.Request) (*http.Response, error) {
						atomic.AddInt32(&calls, 1)
						return regressionResponse(r, 200, io.NopCloser(strings.NewReader(`{}`))), nil
					})}))
					defer c.Close()
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					if preCancelled {
						cancel()
					}
					req, _ := c.newRequest(ctx, http.MethodGet, "/test", nil)
					started := make(chan struct{})
					done := make(chan error, 1)
					go func() {
						close(started)
						var err error
						if raw {
							_, err = c.doRaw(req)
						} else {
							err = c.do(req, nil)
						}
						done <- err
					}()
					<-started
					cancel()
					select {
					case err := <-done:
						if !errors.Is(err, context.Canceled) {
							t.Fatal(err)
						}
					case <-time.After(5 * time.Second):
						t.Fatal("limiter ignored cancellation")
					}
					if atomic.LoadInt32(&calls) != 0 {
						t.Fatalf("transport invoked after cancellation: %d", calls)
					}
				})
			}
		}
	}
}

func TestConcurrentCallsWithCancellation(t *testing.T) {
	var calls int32
	c := newTestClient("fake-key", WithRateLimit(time.Millisecond), WithHTTPClient(&http.Client{Transport: regressionTransport(func(r *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)
		return regressionResponse(r, 200, io.NopCloser(strings.NewReader(`{}`))), nil
	})}))
	defer c.Close()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if i%2 == 0 {
				cancel()
			}
			req, _ := c.newRequest(ctx, http.MethodGet, "/test", nil)
			err := c.do(req, nil)
			if i%2 == 0 && !errors.Is(err, context.Canceled) {
				t.Errorf("canceled call: %v", err)
			}
			if i%2 != 0 && err != nil {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()
	if atomic.LoadInt32(&calls) != 10 {
		t.Fatalf("unexpected transport calls %d", calls)
	}
}

func TestRedirectSecurity(t *testing.T) {
	var destinationCalls int32
	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&destinationCalls, 1)
		if r.Header.Get("X-inFakt-ApiKey") != "" {
			t.Error("key leaked")
		}
		w.WriteHeader(200)
	}))
	defer destination.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cross":
			http.Redirect(w, r, destination.URL, http.StatusFound)
		case "/same":
			http.Redirect(w, r, "/ok", http.StatusFound)
		case "/loop":
			http.Redirect(w, r, "/loop", http.StatusFound)
		case "/ok":
			if r.Header.Get("X-inFakt-ApiKey") != "fake-key" {
				t.Error("same origin key lost")
			}
			w.WriteHeader(200)
		}
	}))
	defer server.Close()
	c := newTestClient("fake-key", WithBaseURL(server.URL))
	defer c.Close()
	for _, path := range []string{"/cross", "/same", "/loop"} {
		req, _ := c.newRequest(context.Background(), http.MethodGet, path, nil)
		err := c.do(req, nil)
		if (path == "/same") != (err == nil) {
			t.Fatalf("path=%s err=%v", path, err)
		}
	}
	if atomic.LoadInt32(&destinationCalls) != 0 {
		t.Fatal("cross-origin destination received request")
	}
	tlsServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, destination.URL, http.StatusFound) }))
	defer tlsServer.Close()
	tlsClient := newTestClient("fake-key", WithBaseURL(tlsServer.URL), WithHTTPClient(tlsServer.Client()))
	defer tlsClient.Close()
	req, _ := tlsClient.newRequest(context.Background(), http.MethodGet, "/test", nil)
	if err := tlsClient.do(req, nil); err == nil {
		t.Fatal("HTTPS downgrade succeeded")
	}
	if atomic.LoadInt32(&destinationCalls) != 0 {
		t.Fatal("downgrade destination received request")
	}
}

func TestRedirectComposesUserPolicyWithoutMutatingClient(t *testing.T) {
	for _, lastResponse := range []bool{false, true} {
		t.Run(map[bool]string{true: "last response", false: "error"}[lastResponse], func(t *testing.T) {
			var callbacks int
			userErr := errors.New("user refused redirect")
			jar, _ := cookiejar.New(nil)
			custom := &http.Client{Timeout: 12 * time.Second, Jar: jar, Transport: http.DefaultTransport, CheckRedirect: func(*http.Request, []*http.Request) error {
				callbacks++
				if lastResponse {
					return http.ErrUseLastResponse
				}
				return userErr
			}}
			callbackPointer := reflect.ValueOf(custom.CheckRedirect).Pointer()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/same", http.StatusFound) }))
			defer server.Close()
			c := newTestClient("fake-key", WithBaseURL(server.URL), WithHTTPClient(custom))
			defer c.Close()
			if c.httpClient == custom || c.httpClient.Timeout != custom.Timeout || c.httpClient.Jar != jar || c.httpClient.Transport != custom.Transport {
				t.Fatal("client configuration not preserved")
			}
			req, _ := c.newRequest(context.Background(), http.MethodGet, "/test", nil)
			err := c.do(req, nil)
			if err == nil {
				t.Fatal("redirect refusal became success")
			}
			if !lastResponse && !errors.Is(err, userErr) {
				t.Fatalf("callback error lost: %v", err)
			}
			if callbacks != 1 || reflect.ValueOf(custom.CheckRedirect).Pointer() != callbackPointer {
				t.Fatal("caller policy changed")
			}
		})
	}
}

func TestSameOrigin(t *testing.T) {
	base, _ := url.Parse("https://EXAMPLE.test/path")
	for target, want := range map[string]bool{"https://example.test:443/other": true, "https://example.test:444/": false, "http://example.test/": false, "https://sub.example.test/": false, "https://other.test/": false} {
		u, _ := url.Parse(target)
		if got := sameOrigin(base, u); got != want {
			t.Errorf("%s got %v want %v", target, got, want)
		}
	}
}

func TestPDFIsNeverRetried(t *testing.T) {
	for _, status := range []int{500, 429} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			calls := 0
			c := newTestClient("fake-key", WithHTTPClient(&http.Client{Transport: regressionTransport(func(r *http.Request) (*http.Response, error) {
				calls++
				return regressionResponse(r, status, io.NopCloser(strings.NewReader(`{"error":"ambiguous render"}`))), nil
			})}))
			defer c.Close()
			_, err := c.Invoices.GetPDF(context.Background(), "test", "", "")
			var apiErr *ErrorResponse
			if !errors.As(err, &apiErr) || apiErr.StatusCode != status || calls != 1 {
				t.Fatalf("err=%v calls=%d", err, calls)
			}
		})
	}
}

func TestRedirectSecurityPrecedesUserPolicy(t *testing.T) {
	var callbacks int
	var destinationCalls int32
	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { atomic.AddInt32(&destinationCalls, 1) }))
	defer destination.Close()
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, destination.URL, http.StatusFound) }))
	defer source.Close()
	c := newTestClient("fake-key", WithBaseURL(source.URL), WithHTTPClient(&http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { callbacks++; return http.ErrUseLastResponse }}))
	defer c.Close()
	req, _ := c.newRequest(context.Background(), http.MethodGet, "/test", nil)
	err := c.do(req, nil)
	if err == nil || callbacks != 0 || atomic.LoadInt32(&destinationCalls) != 0 {
		t.Fatalf("err=%v callbacks=%d requests=%d", err, callbacks, destinationCalls)
	}
	if strings.Contains(err.Error(), "fake-key") {
		t.Fatal("key appeared in error")
	}
}

func TestRedirectLoopBoundWithUserPolicy(t *testing.T) {
	var calls int32
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		http.Redirect(w, r, "/loop", http.StatusFound)
	}))
	defer source.Close()
	c := newTestClient("fake-key", WithBaseURL(source.URL), WithHTTPClient(&http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return nil }}))
	defer c.Close()
	req, _ := c.newRequest(context.Background(), http.MethodGet, "/loop", nil)
	if err := c.do(req, nil); err == nil {
		t.Fatal("loop succeeded")
	}
	if atomic.LoadInt32(&calls) != 10 {
		t.Fatalf("loop requests %d", calls)
	}
}

func TestWriteNetworkFailureIsNotRetried(t *testing.T) {
	calls := 0
	networkErr := errors.New("connection lost after saving")
	c := newTestClient("fake-key", WithHTTPClient(&http.Client{Transport: regressionTransport(func(*http.Request) (*http.Response, error) { calls++; return nil, networkErr })}))
	defer c.Close()
	req, _ := c.newRequest(context.Background(), http.MethodPost, "/test", nil)
	if err := c.do(req, nil); !errors.Is(err, networkErr) {
		t.Fatalf("network error lost: %v", err)
	}
	if calls != 1 {
		t.Fatalf("write retried %d times", calls)
	}
}

func TestRedirectCallbackCannotChangeOrigin(t *testing.T) {
	var calls int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		if r.Header.Get("X-inFakt-ApiKey") != "" {
			t.Error("key leaked via redirect callback URL mutation")
		}
	}))
	defer target.Close()
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ok", http.StatusFound)
	}))
	defer source.Close()
	c := newTestClient("fake-key", WithBaseURL(source.URL), WithHTTPClient(&http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL, _ = url.Parse(target.URL)
			// The callback must not be able to change the reference origin either.
			via[0].URL, _ = url.Parse(target.URL)
			return nil
		},
	}))
	defer c.Close()
	req, _ := c.newRequest(context.Background(), http.MethodGet, "/test", nil)
	if err := c.do(req, nil); err == nil {
		t.Fatal("callback redirected request outside original origin")
	}
	if atomic.LoadInt32(&calls) != 0 {
		t.Fatalf("sent %d requests to foreign origin", calls)
	}
}

func TestCancellationDuringBodyReplay(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	calls := 0
	replayedBody := &regressionBody{Reader: strings.NewReader(`{"x":"y"}`)}
	c := newTestClient("fake-key", WithHTTPClient(&http.Client{Transport: regressionTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Context().Err() != nil {
			t.Error("transport invoked after GetBody canceled context")
		}
		if r.Body != nil {
			_, _ = io.Copy(io.Discard, r.Body)
			_ = r.Body.Close()
		}
		resp := regressionResponse(r, http.StatusTooManyRequests, io.NopCloser(strings.NewReader(`{}`)))
		resp.Header.Set("Retry-After", "0")
		return resp, nil
	})}))
	defer c.Close()
	req, _ := c.newRequest(ctx, http.MethodGet, "/test", map[string]string{"x": "y"})
	req.GetBody = func() (io.ReadCloser, error) {
		cancel()
		return replayedBody, nil
	}
	if err := c.do(req, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled context, got %v", err)
	}
	if calls != 1 || !replayedBody.closed {
		t.Fatalf("requests=%d replayed body closed=%v", calls, replayedBody.closed)
	}
}

func TestRedirectHistoryMutationCannotChangeOrigin(t *testing.T) {
	var targetCalls int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&targetCalls, 1)
		if r.Header.Get("X-inFakt-ApiKey") != "" {
			t.Error("key leaked after redirect history mutation")
		}
	}))
	defer target.Close()
	var sourceCalls int32
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&sourceCalls, 1)
		if r.URL.Path == "/first" {
			http.Redirect(w, r, "/second", http.StatusFound)
			return
		}
		http.Redirect(w, r, target.URL, http.StatusFound)
	}))
	defer source.Close()
	callbacks := 0
	c := newTestClient("fake-key", WithBaseURL(source.URL), WithHTTPClient(&http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			callbacks++
			// Leave the current destination unchanged while corrupting the
			// origin that a subsequent CheckRedirect receives in its history.
			via[0].URL, _ = url.Parse(target.URL)
			return nil
		},
	}))
	defer c.Close()
	req, _ := c.newRequest(context.Background(), http.MethodGet, "/first", nil)
	if err := c.do(req, nil); err == nil {
		t.Fatal("redirect chain escaped the original origin")
	}
	if atomic.LoadInt32(&sourceCalls) != 2 || atomic.LoadInt32(&targetCalls) != 0 || callbacks != 1 {
		t.Fatalf("source requests=%d target requests=%d callbacks=%d", sourceCalls, targetCalls, callbacks)
	}
}
