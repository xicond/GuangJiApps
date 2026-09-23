package e2e

import (
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/gavv/httpexpect/v2"
)

var (
	cachedToken string
	tokenMu     sync.Mutex
)

// getBaseURL retrieves the target host URL for E2E testing from the E2E_BASE_URL environment variable,
// falling back to http://localhost:8080 when unset.
//
// Returns:
//   - string: resolved base URL string.
func getBaseURL() string {
	if url := os.Getenv("E2E_BASE_URL"); url != "" {
		return url
	}
	return "http://localhost:8080"
}

// newExpect initializes an unauthenticated httpexpect client bound to the current test runner.
//
// Parameters:
//   - t: active testing.T pointer.
//
// Returns:
//   - *httpexpect.Expect: initialized HTTP testing expectation client.
func newExpect(t *testing.T) *httpexpect.Expect {
	return httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  getBaseURL(),
		Reporter: httpexpect.NewAssertReporter(t),
	})
}

// getAuthToken authenticates as the default administrator, caching the resulting JWT token
// across test runs protected by a sync.Mutex.
//
// Parameters:
//   - t: active testing.T pointer.
//   - e: active httpexpect client.
//
// Returns:
//   - string: raw JWT token string.
func getAuthToken(t *testing.T, e *httpexpect.Expect) string {
	tokenMu.Lock()
	defer tokenMu.Unlock()

	if cachedToken != "" {
		return cachedToken
	}

	res := e.POST("/login").
		WithHeader("Content-Type", "application/json").
		WithJSON(map[string]string{
			"username": "admin",
			"password": "password",
		}).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	res.ContainsKey("token")
	token := res.Value("token").String().Raw()
	cachedToken = token
	return token
}

// newAuthExpect creates a new expect instance and fetches a Bearer authorization token header.
//
// Parameters:
//   - t: active testing.T pointer.
//
// Returns:
//   - *httpexpect.Expect: initialized HTTP test client.
//   - string: formatted "Bearer <token>" authorization header value.
func newAuthExpect(t *testing.T) (*httpexpect.Expect, string) {
	e := newExpect(t)
	token := getAuthToken(t, e)
	return e, "Bearer " + token
}

// getAuthenticatedExpect initializes an httpexpect client pre-configured to inject
// the administrator Bearer token on every outgoing HTTP request.
//
// Parameters:
//   - t: active testing.T pointer.
//
// Returns:
//   - *httpexpect.Expect: authenticated HTTP request builder client.
func getAuthenticatedExpect(t *testing.T) *httpexpect.Expect {
	e, auth := newAuthExpect(t)
	return e.Builder(func(req *httpexpect.Request) {
		req.WithHeader("Authorization", auth)
	})
}
