package e2e

import (
	"net/http"
	"testing"
)

// TestE2E_Ping verifies application health and responsiveness via the /ping endpoint.
//
// Parameters:
//   - t: active testing.T pointer.
func TestE2E_Ping(t *testing.T) {
	e := newExpect(t)

	e.GET("/ping").
		Expect().
		Status(http.StatusOK).
		JSON().Object().
		Value("message").String().IsEqual("pong")
}

// TestE2E_Login tests authentication flows including successful credential exchange (200),
// malformed request payload (400), and invalid login credentials (401).
//
// Parameters:
//   - t: active testing.T pointer.
func TestE2E_Login(t *testing.T) {
	e := newExpect(t)

	t.Run("success_200", func(t *testing.T) {
		res := e.POST("/login").
			WithJSON(map[string]string{
				"username": "admin",
				"password": "password",
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		res.Value("message").String().IsEqual("login successful")
		res.ContainsKey("token")
		res.Value("token").String().NotEmpty()
		res.ContainsKey("user")
		res.Value("user").Object().Value("username").String().IsEqual("admin")
	})

	t.Run("malformed_json_400", func(t *testing.T) {
		e.POST("/login").
			WithHeader("Content-Type", "application/json").
			WithBytes([]byte("{invalid-json}")).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object().
			ContainsKey("error")
	})

	t.Run("invalid_credentials_401", func(t *testing.T) {
		e.POST("/login").
			WithJSON(map[string]string{
				"username": "wrong_user",
				"password": "wrong_password",
			}).
			Expect().
			Status(http.StatusUnauthorized).
			JSON().Object().
			ContainsKey("error")
	})
}

// TestE2E_ChangePassword verifies password update payload validation and error handling.
//
// Parameters:
//   - t: active testing.T pointer.
func TestE2E_ChangePassword(t *testing.T) {
	e, auth := newAuthExpect(t)

	t.Run("invalid_payload_400", func(t *testing.T) {
		e.POST("/v1/change-password").
			WithHeader("Authorization", auth).
			WithHeader("Content-Type", "application/json").
			WithBytes([]byte("{invalid-json}")).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object().
			ContainsKey("error")
	})
}
