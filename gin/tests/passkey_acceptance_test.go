package tests

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"guangjiapps/gin/internal/api"
	"guangjiapps/gin/internal/config"
	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"
	"guangjiapps/gin/internal/service"
)

// setupAcceptanceServer connects directly to MSSQL database, configures JWT signing keys,
// auto-migrates T_Auth_Passkey, initializes AuthService and the root API router, and starts a test HTTP server.
func setupAcceptanceServer(t *testing.T) (*httptest.Server, *gorm.DB, *service.AuthService, config.Config) {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "server=127.0.0.1;port=1433;user id=sa;password=YourStrong@Passw0rd123;database=guangji;encrypt=disable;"
	}
	db, err := database.Open(dsn)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	db.Exec("DELETE FROM T_Auth_Passkey WHERE LoginId IN (SELECT LoginId FROM T_Login_Mst WHERE Username = 'admin') OR LoginId IN (1, 10, 12, 20, 30)")

	// Ensure LoginId and CredentialId columns exist in MSSQL, and types are VARCHAR(MAX)
	// Recreate table if RefreshToken is still legacy 'text' type (incompatible with = operator in MSSQL)
	// db.Exec(`IF EXISTS (SELECT 1 FROM sys.columns c JOIN sys.types t ON c.user_type_id = t.user_type_id WHERE c.object_id = OBJECT_ID('dbo.T_Auth_Passkey') AND c.name = 'RefreshToken' AND t.name = 'text')
	// BEGIN
	// 	DROP TABLE [dbo].[T_Auth_Passkey];
	// 	CREATE TABLE [dbo].[T_Auth_Passkey] (
	// 		[PasskeyId] INT IDENTITY(1,1) NOT NULL PRIMARY KEY,
	// 		[LoginId] INT NOT NULL,
	// 		[CredentialId] VARCHAR(255) NULL,
	// 		[PublicKey] VARCHAR(MAX) NOT NULL,
	// 		[RefreshToken] VARCHAR(MAX) NOT NULL,
	// 		[RefreshTokenHash] VARCHAR(32) NULL,
	// 		[CreatedAt] DATETIME NULL,
	// 		[LastUsed] DATETIME NULL,
	// 		[ExpiredAt] DATETIME NULL,
	// 		[Status] BIT NULL DEFAULT ((1))
	// 	);
	// 	CREATE NONCLUSTERED INDEX [IX_T_Auth_Passkey_RefreshTokenHash] ON [dbo].[T_Auth_Passkey] ([RefreshTokenHash]);
	// 	CREATE NONCLUSTERED INDEX [IX_T_Auth_Passkey_LoginId] ON [dbo].[T_Auth_Passkey] ([LoginId]);
	// 	CREATE NONCLUSTERED INDEX [IX_T_Auth_Passkey_CredentialId] ON [dbo].[T_Auth_Passkey] ([CredentialId]);
	// END`)

	// Auto-migrate T_Auth_Passkey table in MSSQL
	// if err := db.AutoMigrate(&domain.AuthPasskey{}); err != nil {
	// 	t.Logf("auto-migrate AuthPasskey on mssql: %v", err)
	// }
	// db.Exec("DELETE FROM T_Auth_Passkey WHERE LoginId IN (SELECT LoginId FROM T_Login_Mst WHERE Username = 'admin') OR LoginId IN (1, 12)")

	// Generate ECDSA P-256 private key for JWT and Refresh Token signing
	jwtPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate EC key: %v", err)
	}
	jwtDER, err := x509.MarshalECPrivateKey(jwtPrivKey)
	if err != nil {
		t.Fatalf("failed to marshal EC key: %v", err)
	}
	jwtPEM := string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: jwtDER}))

	cfg := config.Config{
		GinMode:       "test",
		BaseURL:       "",
		JWTSecret:     jwtPEM,
		RefreshSecret: jwtPEM,
		DatabaseDSN:   dsn,
	}

	authService := service.NewAuthService(cfg, db)
	router := api.NewRouter(authService, db, cfg)

	server := httptest.NewServer(router.Engine)
	return server, db, authService, cfg
}

// TestPasskey_Acceptance performs end-to-end acceptance testing for the biometric passkey feature:
// - POST /v1/add-passkey: unauthenticated checks, validation, and successful registration.
// - POST /login-renew: missing token, invalid signature, valid biometric renewal, and old token revocation check.
func TestPasskey_Acceptance(t *testing.T) {
	server, db, authService, _ := setupAcceptanceServer(t)
	defer server.Close()

	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  server.URL,
		Reporter: httpexpect.NewAssertReporter(t),
	})

	// 1. Authenticate with admin / password against live MSSQL
	adminUser, authToken, _, err := authService.Login("admin", "password")
	if err != nil {
		t.Fatalf("failed to login as admin on MSSQL: %v", err)
	}
	authHeader := "Bearer " + authToken

	// Generate authenticator EC key pair for browser biometric simulation
	authPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate authenticator key: %v", err)
	}
	pubKeyDER, _ := x509.MarshalPKIXPublicKey(&authPrivKey.PublicKey)
	pubKeyPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubKeyDER}))

	var issuedRefreshToken string

	// ==========================================
	// 1. Acceptance Tests for POST /v1/add-passkey
	// ==========================================
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

	t.Run("add_passkey_malformed_json_400", func(t *testing.T) {
		e.POST("/v1/add-passkey").
			WithHeader("Authorization", authHeader).
			WithHeader("Content-Type", "application/json").
			WithBytes([]byte("{invalid-json-content")).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object().
			ContainsKey("error")
	})

	t.Run("add_passkey_invalid_challenge_400", func(t *testing.T) {
		badClientData := `{"type":"webauthn.create","challenge":"` + base64.RawURLEncoding.EncodeToString([]byte("mismatched-challenge-token")) + `","origin":"https://app.local"}`
		badClientDataB64 := base64.RawURLEncoding.EncodeToString([]byte(badClientData))

		e.POST("/v1/add-passkey").
			WithHeader("Authorization", authHeader).
			WithHeader("Content-Type", "application/json").
			WithJSON(service.AddPasskeyRequest{
				ID:        "passkey-bad-challenge",
				PublicKey: pubKeyPEM,
				Response: service.WebAuthnAttestationResponse{
					ClientDataJSON: badClientDataB64,
					PublicKey:      pubKeyPEM,
				},
			}).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object().
			ContainsKey("error")
	})

	t.Run("add_passkey_success_200", func(t *testing.T) {
		validClientData := `{"type":"webauthn.create","challenge":"` + base64.RawURLEncoding.EncodeToString([]byte(authToken)) + `","origin":"https://app.local"}`
		validClientDataB64 := base64.RawURLEncoding.EncodeToString([]byte(validClientData))

		res := e.POST("/v1/add-passkey").
			WithHeader("Authorization", authHeader).
			WithHeader("Content-Type", "application/json").
			WithJSON(service.AddPasskeyRequest{
				ID:        "passkey-valid-1",
				RawID:     base64.RawURLEncoding.EncodeToString([]byte("passkey-valid-1")),
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

		// Verify record persisted in DB
		var passkeyRecord domain.AuthPasskey
		err := db.Table("T_Auth_Passkey").Where("RefreshToken = ?", issuedRefreshToken).First(&passkeyRecord).Error
		if err != nil {
			t.Fatalf("failed to find passkey in database: %v", err)
		}
		if passkeyRecord.LoginID != 12 || passkeyRecord.Status == nil || !*passkeyRecord.Status || passkeyRecord.CredentialID == "" {
			t.Fatalf("unexpected passkey state in DB: %+v", passkeyRecord)
		}
	})

	// ==========================================
	// 2. Acceptance Tests for POST /login-renew
	// ==========================================
	t.Run("login_renew_missing_refresh_token_401", func(t *testing.T) {
		e.POST("/login-renew").
			WithHeader("Content-Type", "application/json").
			WithJSON(map[string]interface{}{
				"id": "passkey-valid-1",
			}).
			Expect().
			Status(http.StatusUnauthorized).
			JSON().Object().
			ContainsKey("error")
	})

	t.Run("login_renew_invalid_signature_401", func(t *testing.T) {
		// Sign with a different, unauthorized private key
		wrongPrivKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

		clientDataJSON := `{"type":"webauthn.get","challenge":"` + base64.RawURLEncoding.EncodeToString([]byte(issuedRefreshToken)) + `","origin":"https://app.local"}`
		clientDataB64 := base64.RawURLEncoding.EncodeToString([]byte(clientDataJSON))

		authData := make([]byte, 37)
		rpIDHash := sha256.Sum256([]byte("app.local"))
		copy(authData[0:32], rpIDHash[:])
		authData[32] = 0x05 // UP + UV flags
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
				ID: "passkey-valid-1",
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

	var renewedRefreshToken string

	t.Run("login_renew_success_200", func(t *testing.T) {
		// Sign with the registered, valid private key
		clientDataJSON := `{"type":"webauthn.get","challenge":"` + base64.RawURLEncoding.EncodeToString([]byte(issuedRefreshToken)) + `","origin":"https://app.local"}`
		clientDataB64 := base64.RawURLEncoding.EncodeToString([]byte(clientDataJSON))

		authData := make([]byte, 37)
		rpIDHash := sha256.Sum256([]byte("app.local"))
		copy(authData[0:32], rpIDHash[:])
		authData[32] = 0x05 // UP + UV flags
		authDataB64 := base64.RawURLEncoding.EncodeToString(authData)

		clientDataHash := sha256.Sum256([]byte(clientDataJSON))
		signedData := append(authData, clientDataHash[:]...)
		signedHash := sha256.Sum256(signedData)

		validSigBytes, err := ecdsa.SignASN1(rand.Reader, authPrivKey, signedHash[:])
		if err != nil {
			t.Fatalf("failed to sign assertion: %v", err)
		}
		validSigB64 := base64.RawURLEncoding.EncodeToString(validSigBytes)

		res := e.POST("/login-renew").
			WithHeader("Authorization", "Bearer "+issuedRefreshToken).
			WithHeader("Content-Type", "application/json").
			WithJSON(service.LoginRenewRequest{
				ID: "passkey-valid-1",
				Response: service.WebAuthnAssertionResponse{
					ClientDataJSON:    clientDataB64,
					AuthenticatorData: authDataB64,
					Signature:         validSigB64,
				},
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		// Verify response returns renewed access token and refresh token
		res.ContainsKey("refresh_token")
		res.ContainsKey("token")
		res.ContainsKey("access_token")
		renewedRefreshToken = res.Value("refresh_token").String().NotEmpty().Raw()
		renewedAccessToken := res.Value("token").String().NotEmpty().Raw()
		if renewedAccessToken == "" {
			t.Fatalf("expected non-empty renewed access token")
		}

		if renewedRefreshToken == issuedRefreshToken {
			t.Fatalf("expected renewed refresh token to differ from old token")
		}

		// Verify old refresh token is marked status = 0 (invalidated) in DB
		var oldRecord domain.AuthPasskey
		err = db.Table("T_Auth_Passkey").Where("RefreshToken = ?", issuedRefreshToken).First(&oldRecord).Error
		if err != nil {
			t.Fatalf("failed to lookup old record in DB: %v", err)
		}
		if oldRecord.Status == nil || *oldRecord.Status != false {
			t.Fatalf("expected old refresh token to be invalidated (status=false), got status=%v", oldRecord.Status)
		}

		// Verify new refresh token is active (status = 1) in DB with same public key
		var newRecord domain.AuthPasskey
		err = db.Table("T_Auth_Passkey").Where("RefreshToken = ?", renewedRefreshToken).First(&newRecord).Error
		if err != nil {
			t.Fatalf("failed to lookup new record in DB: %v", err)
		}
		if newRecord.Status == nil || !*newRecord.Status || newRecord.Credential != oldRecord.Credential {
			t.Fatalf("unexpected state for new refresh token in DB: %+v", newRecord)
		}
	})

	t.Run("login_renew_prevent_token_reuse_401", func(t *testing.T) {
		// Attempting to renew again using the invalidated old refresh token must fail
		clientDataJSON := `{"type":"webauthn.get","challenge":"` + base64.RawURLEncoding.EncodeToString([]byte(issuedRefreshToken)) + `","origin":"https://app.local"}`
		clientDataB64 := base64.RawURLEncoding.EncodeToString([]byte(clientDataJSON))

		authData := make([]byte, 37)
		rpIDHash := sha256.Sum256([]byte("app.local"))
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
				ID: "passkey-valid-1",
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

	t.Run("login_renew_chain_renewal_200", func(t *testing.T) {
		// Renewal with the newly issued refresh token must succeed
		clientDataJSON := `{"type":"webauthn.get","challenge":"` + base64.RawURLEncoding.EncodeToString([]byte(renewedRefreshToken)) + `","origin":"https://app.local"}`
		clientDataB64 := base64.RawURLEncoding.EncodeToString([]byte(clientDataJSON))

		authData := make([]byte, 37)
		rpIDHash := sha256.Sum256([]byte("app.local"))
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
			WithHeader("Authorization", "Bearer "+renewedRefreshToken).
			WithHeader("Content-Type", "application/json").
			WithJSON(service.LoginRenewRequest{
				ID: "passkey-valid-1",
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
		secondRenewedToken := res.Value("refresh_token").String().NotEmpty().Raw()
		if secondRenewedToken == renewedRefreshToken {
			t.Fatalf("expected next renewed refresh token to differ")
		}
	})

	// ==========================================
	// 3. Acceptance Tests for POST /login-options
	// ==========================================
	t.Run("login_options_empty_username_400", func(t *testing.T) {
		e.POST("/login-options").
			WithJSON(map[string]string{"username": ""}).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object().ContainsKey("error")
	})

	t.Run("login_options_user_not_found_404", func(t *testing.T) {
		e.POST("/login-options").
			WithJSON(map[string]string{"username": "non_existent_user_9999"}).
			Expect().
			Status(http.StatusNotFound).
			JSON().Object().ContainsKey("error")
	})

	t.Run("login_options_success_single_and_multiple_1_to_many", func(t *testing.T) {
		// Single passkey check
		res := e.POST("/login-options").
			WithHeader("Origin", "https://account.docusign.com").
			WithJSON(map[string]string{"username": "admin"}).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		res.Value("challenge").String().NotEmpty()
		res.Value("rpId").String().IsEqual("account.docusign.com")
		res.Value("userVerification").String().IsEqual("required")
		creds := res.Value("allowCredentials").Array().NotEmpty()
		creds.Length().IsEqual(1)

		firstCred := creds.Element(0).Object()
		firstCred.Value("type").String().IsEqual("public-key")
		firstCred.Value("id").String().NotEmpty()
		firstCred.Value("transports").Array().NotEmpty()

		// 1 to Many test: add a second passkey for the same user (LoginID 12)
		now := time.Now()
		st := true
		secondRefTok := "second-device-refresh-token-999"
		db.Table("T_Auth_Passkey").Create(&domain.AuthPasskey{
			LoginID:          adminUser.ID,
			CredentialID:     "second-device-passkey-id-999",
			Credential:       pubKeyPEM,
			RefreshToken:     &secondRefTok,
			RefreshTokenHash: fmt.Sprintf("%x", md5.Sum([]byte(secondRefTok))),
			CreatedAt:        &now,
			Status:           &st,
		})

		multiRes := e.POST("/login-options").
			WithHeader("Origin", "https://account.docusign.com").
			WithJSON(map[string]string{"username": "admin"}).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		multiCreds := multiRes.Value("allowCredentials").Array()
		multiCreds.Length().IsEqual(2)
	})

	// ==========================================
	// 4. Acceptance Tests for POST /login-passkey
	// ==========================================
	t.Run("login_passkey_invalid_request_400", func(t *testing.T) {
		e.POST("/login-passkey").
			WithJSON(map[string]interface{}{
				"username": "",
			}).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object().ContainsKey("error")
	})

	t.Run("login_passkey_user_not_found_404", func(t *testing.T) {
		e.POST("/login-passkey").
			WithJSON(map[string]interface{}{
				"username": "unknown_user_xyz",
				"credential": map[string]interface{}{
					"id": "passkey-valid-1",
					"response": map[string]interface{}{
						"clientDataJSON":    base64.RawURLEncoding.EncodeToString([]byte(`{"type":"webauthn.get"}`)),
						"authenticatorData": base64.RawURLEncoding.EncodeToString([]byte{0x01}),
						"signature":         base64.RawURLEncoding.EncodeToString([]byte{0x01}),
					},
				},
			}).
			Expect().
			Status(http.StatusNotFound).
			JSON().Object().ContainsKey("error")
	})

	t.Run("login_passkey_success_without_rememberme_200", func(t *testing.T) {
		challenge := "login-passkey-challenge-test"
		clientDataJSON := `{"type":"webauthn.get","challenge":"` + challenge + `","origin":"https://app.local"}`
		clientDataB64 := base64.RawURLEncoding.EncodeToString([]byte(clientDataJSON))

		authData := make([]byte, 37)
		rpIDHash := sha256.Sum256([]byte("app.local"))
		copy(authData[0:32], rpIDHash[:])
		authData[32] = 0x05 // UP + UV
		authDataB64 := base64.RawURLEncoding.EncodeToString(authData)

		clientDataHash := sha256.Sum256([]byte(clientDataJSON))
		signedData := append(authData, clientDataHash[:]...)
		signedHash := sha256.Sum256(signedData)

		sigBytes, err := ecdsa.SignASN1(rand.Reader, authPrivKey, signedHash[:])
		if err != nil {
			t.Fatalf("failed to sign assertion: %v", err)
		}
		sigB64 := base64.RawURLEncoding.EncodeToString(sigBytes)

		var countBefore int64
		db.Table("T_Auth_Passkey").Where("LoginId = ?", adminUser.ID).Count(&countBefore)

		res := e.POST("/login-passkey").
			WithHeader("X-Forwarded-For", "192.168.1.100").
			WithHeader("User-Agent", "Mozilla/5.0 TestPasskeyAgent").
			WithJSON(map[string]interface{}{
				"username":   "admin",
				"rememberme": false,
				"credential": map[string]interface{}{
					"id": "passkey-valid-1",
					"response": map[string]interface{}{
						"clientDataJSON":    clientDataB64,
						"authenticatorData": authDataB64,
						"signature":         sigB64,
					},
				},
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		res.Value("message").String().IsEqual("login successful")
		res.ContainsKey("token")
		res.ContainsKey("access_token")
		res.ContainsKey("user")
		res.ContainsKey("main_menu")
		res.NotContainsKey("refresh_token")
		res.NotContainsKey("public_key")

		var countAfter int64
		db.Table("T_Auth_Passkey").Where("LoginId = ?", adminUser.ID).Count(&countAfter)
		if countAfter != countBefore {
			t.Fatalf("expected passkey count in DB to remain %d without rememberme, got %d", countBefore, countAfter)
		}
	})

	t.Run("login_passkey_success_with_rememberme_ipv6_200", func(t *testing.T) {
		challenge := "login-passkey-rememberme-challenge"
		clientDataJSON := `{"type":"webauthn.get","challenge":"` + challenge + `","origin":"https://app.local"}`
		clientDataB64 := base64.RawURLEncoding.EncodeToString([]byte(clientDataJSON))

		authData := make([]byte, 37)
		rpIDHash := sha256.Sum256([]byte("app.local"))
		copy(authData[0:32], rpIDHash[:])
		authData[32] = 0x05 // UP + UV
		authDataB64 := base64.RawURLEncoding.EncodeToString(authData)

		clientDataHash := sha256.Sum256([]byte(clientDataJSON))
		signedData := append(authData, clientDataHash[:]...)
		signedHash := sha256.Sum256(signedData)

		sigBytes, err := ecdsa.SignASN1(rand.Reader, authPrivKey, signedHash[:])
		if err != nil {
			t.Fatalf("failed to sign assertion: %v", err)
		}
		sigB64 := base64.RawURLEncoding.EncodeToString(sigBytes)

		ipv6Addr := "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
		uaString := "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) GuangJiAcceptance"

		var countBefore int64
		db.Table("T_Auth_Passkey").Where("LoginId = ?", adminUser.ID).Count(&countBefore)

		res := e.POST("/login-passkey").
			WithHeader("X-Forwarded-For", ipv6Addr).
			WithHeader("User-Agent", uaString).
			WithJSON(map[string]interface{}{
				"username":   "admin",
				"rememberme": true,
				"credential": map[string]interface{}{
					"id": "passkey-valid-1",
					"response": map[string]interface{}{
						"clientDataJSON":    clientDataB64,
						"authenticatorData": authDataB64,
						"signature":         sigB64,
					},
				},
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		res.Value("message").String().IsEqual("login successful")
		res.ContainsKey("token")
		res.ContainsKey("access_token")
		res.ContainsKey("user")
		res.ContainsKey("main_menu")
		res.ContainsKey("refresh_token")
		res.ContainsKey("public_key")

		returnedRefresh := res.Value("refresh_token").String().NotEmpty().Raw()

		// Verify that a new record was added to T_Auth_Passkey without invalidating others
		var countAfter int64
		db.Table("T_Auth_Passkey").Where("LoginId = ?", adminUser.ID).Count(&countAfter)
		if countAfter != countBefore+1 {
			t.Fatalf("expected passkey count in DB to increase by 1, before=%d, after=%d", countBefore, countAfter)
		}

		// Verify the new record in DB has the IPv6 and UserAgent
		var newPasskey domain.AuthPasskey
		err = db.Table("T_Auth_Passkey").Where("RefreshToken = ?", returnedRefresh).First(&newPasskey).Error
		if err != nil {
			t.Fatalf("failed to find newly created passkey in DB: %v", err)
		}
		if newPasskey.IP == nil || *newPasskey.IP != ipv6Addr {
			t.Fatalf("expected IP %s, got %v", ipv6Addr, newPasskey.IP)
		}
		if newPasskey.UserAgent == nil || *newPasskey.UserAgent != uaString {
			t.Fatalf("expected UserAgent %s, got %v", uaString, newPasskey.UserAgent)
		}
		if newPasskey.Status == nil || !*newPasskey.Status {
			t.Fatalf("expected new passkey to be active (Status=true)")
		}

		// Verify existing passkey (passkey-valid-1) was NOT invalidated
		var existing domain.AuthPasskey
		err = db.Table("T_Auth_Passkey").
			Where("(CredentialId = ? OR CredentialId = ?) AND (Status = 1 OR Status IS NULL)", "passkey-valid-1", base64.RawURLEncoding.EncodeToString([]byte("passkey-valid-1"))).
			Take(&existing).Error
		if err != nil {
			t.Fatalf("failed to find existing passkey: %v", err)
		}
		if existing.Status == nil || !*existing.Status {
			t.Fatalf("expected existing passkey to remain active, but status is %v", existing.Status)
		}
	})
}

// TestPasskey_FullCycleAndTimeTravel proves the complete lifecycle:
// 1. Initial authentication via POST /login returning an access token.
// 2. Accessing a protected endpoint with the login access token (200 OK).
// 3. Registering biometric passkey via POST /v1/add-passkey issuing a refresh token.
// 4. Time travel: Access token expires (24h later); accessing protected endpoint fails with 401.
// 5. Biometric session recovery via POST /login-renew issuing renewed access & refresh tokens.
// 6. Access restored: protected endpoint succeeds with the renewed access token (200 OK).
// 7. Old refresh token replay rejection (401 Unauthorized).
// 8. Time travel: Refresh token expires (30d later); POST /login-renew fails with 401.
// 9. Continuous session rotation: renewing again with latest active refresh token (200 OK).
func TestPasskey_FullCycleAndTimeTravel(t *testing.T) {
	server, db, _, cfg := setupAcceptanceServer(t)
	defer server.Close()

	e := httpexpect.WithConfig(httpexpect.Config{
		BaseURL:  server.URL,
		Reporter: httpexpect.NewAssertReporter(t),
	})

	authPrivKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate authenticator key: %v", err)
	}
	pubKeyDER, _ := x509.MarshalPKIXPublicKey(&authPrivKey.PublicKey)
	pubKeyPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubKeyDER}))

	var (
		loginToken     string
		initialRefresh string
		renewedToken   string
		renewedRefresh string
	)

	// Step 1: POST /login credentials
	t.Run("step1_login_full_cycle", func(t *testing.T) {
		res := e.POST("/login").
			WithHeader("Content-Type", "application/json").
			WithJSON(map[string]string{
				"username": "admin",
				"password": "password",
			}).
			Expect().
			Status(http.StatusOK).
			JSON().Object()

		res.Value("message").String().IsEqual("login successful")
		res.ContainsKey("token")
		loginToken = res.Value("token").String().NotEmpty().Raw()
	})

	// Step 2: Access protected route with initial access token
	t.Run("step2_access_protected_with_login_token", func(t *testing.T) {
		e.GET("/v1/umats").
			WithHeader("Authorization", "Bearer "+loginToken).
			Expect().
			Status(http.StatusOK)
	})

	// Step 3: Register biometric passkey (Remember Me flow)
	t.Run("step3_register_passkey_remember_me", func(t *testing.T) {
		clientData := `{"type":"webauthn.create","challenge":"` + base64.RawURLEncoding.EncodeToString([]byte(loginToken)) + `","origin":"https://guangjiapp.local"}`
		clientDataB64 := base64.RawURLEncoding.EncodeToString([]byte(clientData))

		res := e.POST("/v1/add-passkey").
			WithHeader("Authorization", "Bearer "+loginToken).
			WithHeader("Content-Type", "application/json").
			WithJSON(service.AddPasskeyRequest{
				ID:        "full-cycle-passkey-1",
				RawID:     base64.RawURLEncoding.EncodeToString([]byte("full-cycle-passkey-1")),
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

	// Step 4: Time Travel - simulate access token expiration (24h in future)
	t.Run("step4_time_travel_access_token_expired", func(t *testing.T) {
		vKey, method, err := cfg.GetJWTSigningKey()
		if err != nil {
			t.Fatalf("failed to load signing key: %v", err)
		}
		// Create expired access token (1 hour in the past)
		expiredTokenObj := jwt.NewWithClaims(method, jwt.MapClaims{
			"sub": 1,
			"exp": time.Now().Add(-1 * time.Hour).Unix(),
		})
		expiredTokenStr, err := expiredTokenObj.SignedString(vKey)
		if err != nil {
			t.Fatalf("failed to sign expired token: %v", err)
		}

		// Accessing protected endpoint with expired token must return 401
		e.GET("/v1/umats").
			WithHeader("Authorization", "Bearer "+expiredTokenStr).
			Expect().
			Status(http.StatusUnauthorized).
			JSON().Object().
			Value("error").String().IsEqual("invalid token")
	})

	// Step 5: Biometric Renewal via POST /login-renew
	t.Run("step5_biometric_renewal_recovers_session", func(t *testing.T) {
		clientDataJSON := `{"type":"webauthn.get","challenge":"` + base64.RawURLEncoding.EncodeToString([]byte(initialRefresh)) + `","origin":"https://guangjiapp.local"}`
		clientDataB64 := base64.RawURLEncoding.EncodeToString([]byte(clientDataJSON))

		authData := make([]byte, 37)
		rpIDHash := sha256.Sum256([]byte("guangjiapp.local"))
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
				ID: "full-cycle-passkey-1",
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
			t.Fatalf("expected renewed access token to differ from initial login token")
		}
		if renewedRefresh == initialRefresh {
			t.Fatalf("expected renewed refresh token to differ from initial refresh token")
		}
	})

	// Step 6: Access protected endpoint with renewed access token
	t.Run("step6_access_restored_with_renewed_token", func(t *testing.T) {
		e.GET("/v1/umats").
			WithHeader("Authorization", "Bearer "+renewedToken).
			Expect().
			Status(http.StatusOK)
	})

	// Step 7: Verify replay attack prevention on old refresh token
	t.Run("step7_prevent_old_refresh_token_replay", func(t *testing.T) {
		clientDataJSON := `{"type":"webauthn.get","challenge":"` + base64.RawURLEncoding.EncodeToString([]byte(initialRefresh)) + `","origin":"https://guangjiapp.local"}`
		clientDataB64 := base64.RawURLEncoding.EncodeToString([]byte(clientDataJSON))

		authData := make([]byte, 37)
		rpIDHash := sha256.Sum256([]byte("guangjiapp.local"))
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
				ID: "full-cycle-passkey-1",
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

	// Step 8: Time travel - simulate refresh token expiration (30d in future)
	t.Run("step8_time_travel_refresh_token_expired", func(t *testing.T) {
		vKey, method, err := cfg.GetJWTRefreshSigningKey()
		if err != nil {
			t.Fatalf("failed to load refresh key: %v", err)
		}
		// Create expired refresh token (1 hour in the past)
		expiredRefreshObj := jwt.NewWithClaims(method, jwt.MapClaims{
			"sub":  1,
			"exp":  time.Now().Add(-1 * time.Hour).Unix(),
			"type": "refresh",
		})
		expiredRefreshStr, err := expiredRefreshObj.SignedString(vKey)
		if err != nil {
			t.Fatalf("failed to sign expired refresh token: %v", err)
		}

		// Also persist in DB as expired
		past := time.Now().Add(-1 * time.Hour)
		status := true
		tokenHash := fmt.Sprintf("%x", md5.Sum([]byte(expiredRefreshStr)))
		db.Table("T_Auth_Passkey").Create(&domain.AuthPasskey{
			LoginID:          12,
			CredentialID:     "expired-cred-id-step8",
			Credential:       pubKeyPEM,
			RefreshToken:     &expiredRefreshStr,
			RefreshTokenHash: tokenHash,
			CreatedAt:        &past,
			LastUsed:         &past,
			ExpiredAt:        &past,
			Status:           &status,
		})

		e.POST("/login-renew").
			WithHeader("Authorization", "Bearer "+expiredRefreshStr).
			WithHeader("Content-Type", "application/json").
			WithJSON(service.LoginRenewRequest{
				ID: "full-cycle-passkey-1",
			}).
			Expect().
			Status(http.StatusUnauthorized).
			JSON().Object().
			Value("error").String().IsEqual("refresh token expired")
	})

	// Step 9: Continuous chain rotation with renewed refresh token
	t.Run("step9_continuous_chain_renewal", func(t *testing.T) {
		clientDataJSON := `{"type":"webauthn.get","challenge":"` + base64.RawURLEncoding.EncodeToString([]byte(renewedRefresh)) + `","origin":"https://guangjiapp.local"}`
		clientDataB64 := base64.RawURLEncoding.EncodeToString([]byte(clientDataJSON))

		authData := make([]byte, 37)
		rpIDHash := sha256.Sum256([]byte("guangjiapp.local"))
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
				ID: "full-cycle-passkey-1",
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
