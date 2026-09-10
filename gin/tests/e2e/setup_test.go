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

func getBaseURL() string {
	if url := os.Getenv("E2E_BASE_URL"); url != "" {
		return url
	}
	return "http://localhost:8080"
}

func newExpect(t *testing.T) *httpexpect.Expect {
	return httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  getBaseURL(),
		Reporter: httpexpect.NewAssertReporter(t),
	})
}

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

func newAuthExpect(t *testing.T) (*httpexpect.Expect, string) {
	e := newExpect(t)
	token := getAuthToken(t, e)
	return e, "Bearer " + token
}

func getAuthenticatedExpect(t *testing.T) *httpexpect.Expect {
	e, auth := newAuthExpect(t)
	return e.Builder(func(req *httpexpect.Request) {
		req.WithHeader("Authorization", auth)
	})
}
