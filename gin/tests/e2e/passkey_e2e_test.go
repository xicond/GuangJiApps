package e2e

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"net/http"
	"strings"
	"testing"
	"time"

	"guangjiapps/gin/internal/service"
)

// TestE2E_Passkey validates the full end-to-end integration of WebAuthn passkey registration
// (POST /v1/add-passkey) and biometric token renewal (POST /login-renew).
//
// Parameters:
//   - t: active testing.T pointer.
func TestE2E_Passkey(t *testing.T) {
	e, authHeader := newAuthExpect(t)

	// Extract raw bearer token from authHeader
	bearerToken := strings.TrimPrefix(authHeader, "Bearer ")

	// Generate authenticator EC key pair for browser biometric simulation
	authPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate authenticator key: %v", err)
	}
	pubKeyDER, err := x509.MarshalPKIXPublicKey(&authPrivKey.PublicKey)
	if err != nil {
		t.Fatalf("failed to marshal public key: %v", err)
	}
	pubKeyPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubKeyDER}))

	var issuedRefreshToken string

	// 1. POST /v1/add-passkey unauthenticated
	t.Run("add_passkey_unauthorized_401", func(t *testing.T) {
		e.POST("/v1/add-passkey").
			WithHeader("Content-Type", "application/json").
			WithJSON(map[string]interface{}{
				"id": "sample-passkey",
			}).
			Expect().
			Status(http.StatusUnauthorized).
			JSON().Object().
			ContainsKey("error")
	})

	// 2. POST /v1/add-passkey malformed json
	t.Run("add_passkey_malformed_json_400", func(t *testing.T) {
		e.POST("/v1/add-passkey").
			WithHeader("Authorization", authHeader).
			WithHeader("Content-Type", "application/json").
			WithBytes([]byte("{bad-json")).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object().
			ContainsKey("error")
	})

	// 3. POST /v1/add-passkey successful registration
	t.Run("add_passkey_success_200", func(t *testing.T) {
		validClientData := `{"type":"webauthn.create","challenge":"` + base64.RawURLEncoding.EncodeToString([]byte(bearerToken)) + `","origin":"http://localhost:8080"}`
		validClientDataB64 := base64.RawURLEncoding.EncodeToString([]byte(validClientData))

		res := e.POST("/v1/add-passkey").
			WithHeader("Authorization", authHeader).
			WithHeader("Content-Type", "application/json").
			WithJSON(service.AddPasskeyRequest{
				ID:        "passkey-e2e-1",
				RawID:     base64.RawURLEncoding.EncodeToString([]byte("passkey-e2e-1")),
				Type:      "public-key",
				PublicKey: pubKeyPEM,
				Response: service.WebAuthnAttestationResponse{
					ClientDataJSON: validClientDataB64,
					PublicKey:      pubKeyPEM,
				},
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		res.Value("message").String().IsEqual("passkey added successfully")
		res.ContainsKey("refresh_token")
		issuedRefreshToken = res.Value("refresh_token").String().NotEmpty().Raw()
	})

	// 4. POST /login-renew missing refresh token
	t.Run("login_renew_missing_token_401", func(t *testing.T) {
		e.POST("/login-renew").
			WithHeader("Content-Type", "application/json").
			WithJSON(map[string]interface{}{
				"id": "passkey-e2e-1",
			}).
			Expect().
			Status(http.StatusUnauthorized).
			JSON().Object().
			ContainsKey("error")
	})

	// 5. POST /login-renew invalid signature
	t.Run("login_renew_invalid_signature_401", func(t *testing.T) {
		if issuedRefreshToken == "" {
			t.Skip("skipping because add-passkey did not issue a refresh token")
		}
		wrongPrivKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

		clientDataJSON := `{"type":"webauthn.get","challenge":"` + base64.RawURLEncoding.EncodeToString([]byte(issuedRefreshToken)) + `","origin":"http://localhost:8080"}`
		clientDataB64 := base64.RawURLEncoding.EncodeToString([]byte(clientDataJSON))

		authData := make([]byte, 37)
		rpIDHash := sha256.Sum256([]byte("localhost"))
		copy(authData[0:32], rpIDHash[:])
		authData[32] = 0x05
		authDataB64 := base64.RawURLEncoding.EncodeToString(authData)

		clientDataHash := sha256.Sum256([]byte(clientDataJSON))
		signedData := append(authData, clientDataHash[:]...)
		signedHash := sha256.Sum256(signedData)

		badSigBytes, _ := ecdsa.SignASN1(rand.Reader, wrongPrivKey, signedHash[:])
		badSigB64 := base64.RawURLEncoding.EncodeToString(badSigBytes)

		e.POST("/login-renew").
			WithHeader("Authorization", "Bearer "+issuedRefreshToken).
			WithHeader("Content-Type", "application/json").
			WithJSON(service.LoginRenewRequest{
				ID: "passkey-e2e-1",
				Response: service.WebAuthnAssertionResponse{
					ClientDataJSON:    clientDataB64,
					AuthenticatorData: authDataB64,
					Signature:         badSigB64,
				},
			}).
			Expect().
			Status(http.StatusUnauthorized).
			JSON().Object().
			ContainsKey("error")
	})

	// 6. POST /login-renew successful biometric renewal
	t.Run("login_renew_success_200", func(t *testing.T) {
		if issuedRefreshToken == "" {
			t.Skip("skipping because add-passkey did not issue a refresh token")
		}

		clientDataJSON := `{"type":"webauthn.get","challenge":"` + base64.RawURLEncoding.EncodeToString([]byte(issuedRefreshToken)) + `","origin":"http://localhost:8080"}`
		clientDataB64 := base64.RawURLEncoding.EncodeToString([]byte(clientDataJSON))

		authData := make([]byte, 37)
		rpIDHash := sha256.Sum256([]byte("localhost"))
		copy(authData[0:32], rpIDHash[:])
		authData[32] = 0x05
		authDataB64 := base64.RawURLEncoding.EncodeToString(authData)

		clientDataHash := sha256.Sum256([]byte(clientDataJSON))
		signedData := append(authData, clientDataHash[:]...)
		signedHash := sha256.Sum256(signedData)

		sigBytes, err := ecdsa.SignASN1(rand.Reader, authPrivKey, signedHash[:])
		if err != nil {
			t.Fatalf("failed to sign assertion: %v", err)
		}
		sigB64 := base64.RawURLEncoding.EncodeToString(sigBytes)

		res := e.POST("/login-renew").
			WithHeader("Authorization", "Bearer "+issuedRefreshToken).
			WithHeader("Content-Type", "application/json").
			WithJSON(service.LoginRenewRequest{
				ID: "passkey-e2e-1",
				Response: service.WebAuthnAssertionResponse{
					ClientDataJSON:    clientDataB64,
					AuthenticatorData: authDataB64,
					Signature:         sigB64,
				},
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		res.ContainsKey("refresh_token")
		res.Value("refresh_token").String().NotEmpty()
		res.ContainsKey("token")
		res.Value("token").String().NotEmpty()
	})

	// 7. POST /login-renew old token reuse rejected
	t.Run("login_renew_prevent_token_reuse_401", func(t *testing.T) {
		if issuedRefreshToken == "" {
			t.Skip("skipping because add-passkey did not issue a refresh token")
		}

		clientDataJSON := `{"type":"webauthn.get","challenge":"` + base64.RawURLEncoding.EncodeToString([]byte(issuedRefreshToken)) + `","origin":"http://localhost:8080"}`
		clientDataB64 := base64.RawURLEncoding.EncodeToString([]byte(clientDataJSON))

		authData := make([]byte, 37)
		rpIDHash := sha256.Sum256([]byte("localhost"))
		copy(authData[0:32], rpIDHash[:])
		authData[32] = 0x05
		authDataB64 := base64.RawURLEncoding.EncodeToString(authData)

		clientDataHash := sha256.Sum256([]byte(clientDataJSON))
		signedData := append(authData, clientDataHash[:]...)
		signedHash := sha256.Sum256(signedData)

		sigBytes, _ := ecdsa.SignASN1(rand.Reader, authPrivKey, signedHash[:])
		sigB64 := base64.RawURLEncoding.EncodeToString(sigBytes)

		e.POST("/login-renew").
			WithHeader("Authorization", "Bearer "+issuedRefreshToken).
			WithHeader("Content-Type", "application/json").
			WithJSON(service.LoginRenewRequest{
				ID: "passkey-e2e-1",
				Response: service.WebAuthnAssertionResponse{
					ClientDataJSON:    clientDataB64,
					AuthenticatorData: authDataB64,
					Signature:         sigB64,
				},
			}).
			Expect().
			Status(http.StatusUnauthorized).
			JSON().Object().
			ContainsKey("error")
	})
}

// TestE2E_Passkey_FullCycleAndTimeTravel validates the full cycle from live login,
// protected resource access, passkey registration, biometric renewal, access restoration,
// old token replay prevention, and continuous rotation.
func TestE2E_Passkey_FullCycleAndTimeTravel(t *testing.T) {
	e := newExpect(t)

	// Check if server is running before executing live E2E test
	checkClient := http.Client{Timeout: 500 * time.Millisecond}
	resp, err := checkClient.Get(getBaseURL())
	if err != nil {
		t.Skipf("skipping E2E test: server at %s not reachable: %v", getBaseURL(), err)
		return
	}
	_ = resp.Body.Close()

	authPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate authenticator key: %v", err)
	}
	pubKeyDER, _ := x509.MarshalPKIXPublicKey(&authPrivKey.PublicKey)
	pubKeyPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubKeyDER}))

	var (
		loginToken       string
		initialRefresh   string
		renewedToken     string
		renewedRefresh   string
	)

	// Step 1: POST /login credentials
	t.Run("step1_login_credentials", func(t *testing.T) {
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
		loginToken = res.Value("token").String().NotEmpty().Raw()
	})

	if loginToken == "" {
		t.Skip("skipping remaining steps because login did not return a token")
	}

	// Step 2: Access protected route with initial access token
	t.Run("step2_access_protected_route", func(t *testing.T) {
		e.GET("/v1/umats").
			WithHeader("Authorization", "Bearer "+loginToken).
			Expect().
			Status(http.StatusOK)
	})

	// Step 3: Register passkey via POST /v1/add-passkey (Remember Me flow)
	t.Run("step3_register_passkey_remember_me", func(t *testing.T) {
		clientData := `{"type":"webauthn.create","challenge":"` + base64.RawURLEncoding.EncodeToString([]byte(loginToken)) + `","origin":"http://localhost:8080"}`
		clientDataB64 := base64.RawURLEncoding.EncodeToString([]byte(clientData))

		res := e.POST("/v1/add-passkey").
			WithHeader("Authorization", "Bearer "+loginToken).
			WithHeader("Content-Type", "application/json").
			WithJSON(service.AddPasskeyRequest{
				ID:        "e2e-full-cycle-1",
				RawID:     base64.RawURLEncoding.EncodeToString([]byte("e2e-full-cycle-1")),
				Type:      "public-key",
				PublicKey: pubKeyPEM,
				Response: service.WebAuthnAttestationResponse{
					ClientDataJSON: clientDataB64,
					PublicKey:      pubKeyPEM,
				},
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		res.Value("message").String().IsEqual("passkey added successfully")
		res.ContainsKey("refresh_token")
		initialRefresh = res.Value("refresh_token").String().NotEmpty().Raw()
	})

	if initialRefresh == "" {
		t.Skip("skipping renewal tests because passkey registration did not return a refresh token")
	}

	// Step 4: Biometric renewal via POST /login-renew
	t.Run("step4_biometric_renewal", func(t *testing.T) {
		clientDataJSON := `{"type":"webauthn.get","challenge":"` + base64.RawURLEncoding.EncodeToString([]byte(initialRefresh)) + `","origin":"http://localhost:8080"}`
		clientDataB64 := base64.RawURLEncoding.EncodeToString([]byte(clientDataJSON))

		authData := make([]byte, 37)
		rpIDHash := sha256.Sum256([]byte("localhost"))
		copy(authData[0:32], rpIDHash[:])
		authData[32] = 0x05
		authDataB64 := base64.RawURLEncoding.EncodeToString(authData)

		clientDataHash := sha256.Sum256([]byte(clientDataJSON))
		signedData := append(authData, clientDataHash[:]...)
		signedHash := sha256.Sum256(signedData)

		sigBytes, err := ecdsa.SignASN1(rand.Reader, authPrivKey, signedHash[:])
		if err != nil {
			t.Fatalf("failed to sign assertion: %v", err)
		}
		sigB64 := base64.RawURLEncoding.EncodeToString(sigBytes)

		res := e.POST("/login-renew").
			WithHeader("Authorization", "Bearer "+initialRefresh).
			WithHeader("Content-Type", "application/json").
			WithJSON(service.LoginRenewRequest{
				ID: "e2e-full-cycle-1",
				Response: service.WebAuthnAssertionResponse{
					ClientDataJSON:    clientDataB64,
					AuthenticatorData: authDataB64,
					Signature:         sigB64,
				},
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		res.ContainsKey("token")
		res.ContainsKey("refresh_token")
		renewedToken = res.Value("token").String().NotEmpty().Raw()
		renewedRefresh = res.Value("refresh_token").String().NotEmpty().Raw()

		if renewedToken == loginToken {
			t.Fatalf("expected renewed access token to differ from login token")
		}
		if renewedRefresh == initialRefresh {
			t.Fatalf("expected renewed refresh token to differ from initial refresh token")
		}
	})

	// Step 5: Verify restored access using the renewed access token
	t.Run("step5_access_protected_with_renewed_token", func(t *testing.T) {
		e.GET("/v1/umats").
			WithHeader("Authorization", "Bearer "+renewedToken).
			Expect().
			Status(http.StatusOK)
	})

	// Step 6: Verify replay attack rejection on old refresh token
	t.Run("step6_prevent_old_token_replay", func(t *testing.T) {
		clientDataJSON := `{"type":"webauthn.get","challenge":"` + base64.RawURLEncoding.EncodeToString([]byte(initialRefresh)) + `","origin":"http://localhost:8080"}`
		clientDataB64 := base64.RawURLEncoding.EncodeToString([]byte(clientDataJSON))

		authData := make([]byte, 37)
		rpIDHash := sha256.Sum256([]byte("localhost"))
		copy(authData[0:32], rpIDHash[:])
		authData[32] = 0x05
		authDataB64 := base64.RawURLEncoding.EncodeToString(authData)

		clientDataHash := sha256.Sum256([]byte(clientDataJSON))
		signedData := append(authData, clientDataHash[:]...)
		signedHash := sha256.Sum256(signedData)

		sigBytes, _ := ecdsa.SignASN1(rand.Reader, authPrivKey, signedHash[:])
		sigB64 := base64.RawURLEncoding.EncodeToString(sigBytes)

		e.POST("/login-renew").
			WithHeader("Authorization", "Bearer "+initialRefresh).
			WithHeader("Content-Type", "application/json").
			WithJSON(service.LoginRenewRequest{
				ID: "e2e-full-cycle-1",
				Response: service.WebAuthnAssertionResponse{
					ClientDataJSON:    clientDataB64,
					AuthenticatorData: authDataB64,
					Signature:         sigB64,
				},
			}).
			Expect().
			Status(http.StatusUnauthorized).
			JSON().Object().
			Value("error").String().IsEqual("refresh token not found or revoked")
	})

	// Step 7: Continuous chain rotation with renewed refresh token
	t.Run("step7_continuous_chain_renewal", func(t *testing.T) {
		clientDataJSON := `{"type":"webauthn.get","challenge":"` + base64.RawURLEncoding.EncodeToString([]byte(renewedRefresh)) + `","origin":"http://localhost:8080"}`
		clientDataB64 := base64.RawURLEncoding.EncodeToString([]byte(clientDataJSON))

		authData := make([]byte, 37)
		rpIDHash := sha256.Sum256([]byte("localhost"))
		copy(authData[0:32], rpIDHash[:])
		authData[32] = 0x05
		authDataB64 := base64.RawURLEncoding.EncodeToString(authData)

		clientDataHash := sha256.Sum256([]byte(clientDataJSON))
		signedData := append(authData, clientDataHash[:]...)
		signedHash := sha256.Sum256(signedData)

		sigBytes, err := ecdsa.SignASN1(rand.Reader, authPrivKey, signedHash[:])
		if err != nil {
			t.Fatalf("failed to sign assertion: %v", err)
		}
		sigB64 := base64.RawURLEncoding.EncodeToString(sigBytes)

		res := e.POST("/login-renew").
			WithHeader("Authorization", "Bearer "+renewedRefresh).
			WithHeader("Content-Type", "application/json").
			WithJSON(service.LoginRenewRequest{
				ID: "e2e-full-cycle-1",
				Response: service.WebAuthnAssertionResponse{
					ClientDataJSON:    clientDataB64,
					AuthenticatorData: authDataB64,
					Signature:         sigB64,
				},
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		res.ContainsKey("token")
		res.ContainsKey("refresh_token")
		thirdRefresh := res.Value("refresh_token").String().NotEmpty().Raw()
		if thirdRefresh == renewedRefresh {
			t.Fatalf("expected newly rotated refresh token to differ")
		}
	})
}

