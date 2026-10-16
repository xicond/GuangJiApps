package service

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"
	"testing"
	"time"

	"guangjiapps/gin/internal/config"
	"guangjiapps/gin/internal/database"
	"guangjiapps/gin/internal/domain"
)

func TestAuthService(t *testing.T) {
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate EC key: %v", err)
	}
	ecDER, _ := x509.MarshalECPrivateKey(ecKey)
	ecPEM := string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: ecDER}))

	cfg := config.Config{
		JWTSecret:     ecPEM,
		RefreshSecret: ecPEM,
	}
	db, err := database.Open("")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	authSvc := NewAuthService(cfg, db)

	// 1. Password encryption test (legacy MD5 Base64)
	encrypted := EncryptPassword("123456")
	if encrypted == "" {
		t.Errorf("expected non-empty encrypted password")
	}

	// 2. JWT Generation test
	token, err := authSvc.GenerateToken(1)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Errorf("expected non-empty token")
	}

	// 3. Login with empty parameters
	_, _, _, err = authSvc.Login("", "")
	if err == nil {
		t.Errorf("expected error for empty credentials, got nil")
	}

	// 4. ChangePassword validation tests
	err = authSvc.ChangePassword(1, "", "", "newpass")
	if err == nil {
		t.Errorf("expected error for empty old password, got nil")
	}

	err = authSvc.ChangePassword(1, "", "samepass", "samepass")
	if err == nil {
		t.Errorf("expected error when old and new password match, got nil")
	}

	err = authSvc.ChangePassword(99999, "nonexistent", "oldpass", "newpass")
	if err == nil {
		t.Errorf("expected user not found error for non-existent user, got nil")
	}
}

func TestAuthService_Passkey(t *testing.T) {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "server=127.0.0.1;port=1433;user id=sa;password=YourStrong@Passw0rd123;database=guangji;encrypt=disable;"
	}
	db, err := database.Open(dsn)
	if err != nil {
		t.Fatalf("failed to open mssql test db: %v", err)
	}
	_ = database.AutoMigrate(db)

	var adminUser domain.Admin
	db.Table("T_Login_Mst").Where("Username = 'admin'").First(&adminUser)
	adminID := adminUser.ID
	if adminID == 0 {
		adminID = 12
	}

	db.Exec("DELETE FROM T_Auth_Passkey WHERE LoginId IN (SELECT LoginId FROM T_Login_Mst WHERE Username = 'admin') OR LoginId IN (1, 12)")
	defer db.Exec("DELETE FROM T_Auth_Passkey WHERE LoginId IN (SELECT LoginId FROM T_Login_Mst WHERE Username = 'admin') OR LoginId IN (1, 12)")

	// Generate ECDSA private key for signing JWTs in tests
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate EC key: %v", err)
	}
	ecDER, _ := x509.MarshalECPrivateKey(ecKey)
	ecPEM := string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: ecDER}))

	cfg := config.Config{
		JWTSecret:     ecPEM,
		RefreshSecret: ecPEM,
	}

	authSvc := NewAuthService(cfg, db)

	// 1. GenerateRefreshToken
	rt, exp, err := authSvc.GenerateRefreshToken(adminID)
	if err != nil || rt == "" || exp.Before(time.Now()) {
		t.Fatalf("failed GenerateRefreshToken: %v", err)
	}

	// 2. AddPasskey test
	bearerToken := "sample-bearer-token-123"
	pubKeyDER, _ := x509.MarshalPKIXPublicKey(&ecKey.PublicKey)
	pubKeyPEM := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubKeyDER}))

	clientDataJSON := `{"type":"webauthn.create","challenge":"` + base64.RawURLEncoding.EncodeToString([]byte(bearerToken)) + `","origin":"https://app.local"}`
	clientDataB64 := base64.RawURLEncoding.EncodeToString([]byte(clientDataJSON))

	addReq := AddPasskeyRequest{
		ID:        "passkey-1",
		PublicKey: pubKeyPEM,
		Response: WebAuthnAttestationResponse{
			ClientDataJSON: clientDataB64,
			PublicKey:      pubKeyPEM,
		},
	}

	newRT, err := authSvc.AddPasskey(adminID, addReq, bearerToken, "127.0.0.1", "Go-Test-Agent/1.0")
	if err != nil {
		t.Fatalf("AddPasskey failed: %v", err)
	}
	if newRT == "" {
		t.Fatalf("expected non-empty refresh token from AddPasskey")
	}

	// Verify in DB
	var passkey domain.AuthPasskey
	err = db.Table("T_Auth_Passkey").Where("RefreshToken = ?", newRT).First(&passkey).Error
	if err != nil {
		t.Fatalf("failed to find passkey in DB: %v", err)
	}
	if passkey.LoginID != adminID || passkey.Status == nil || !*passkey.Status || passkey.CredentialID == "" {
		t.Fatalf("unexpected passkey record: %+v", passkey)
	}

	// 3. LoginRenew test
	renewClientDataJSON := `{"type":"webauthn.get","challenge":"` + base64.RawURLEncoding.EncodeToString([]byte(newRT)) + `","origin":"https://app.local"}`
	renewClientDataB64 := base64.RawURLEncoding.EncodeToString([]byte(renewClientDataJSON))

	authData := make([]byte, 37)
	rpIDHash := sha256.Sum256([]byte("app.local"))
	copy(authData[0:32], rpIDHash[:])
	authData[32] = 0x01 // UP
	authDataB64 := base64.RawURLEncoding.EncodeToString(authData)

	renewClientDataHash := sha256.Sum256([]byte(renewClientDataJSON))
	signedData := append(authData, renewClientDataHash[:]...)
	signedHash := sha256.Sum256(signedData)

	sigBytes, err := ecdsa.SignASN1(rand.Reader, ecKey, signedHash[:])
	if err != nil {
		t.Fatalf("failed to sign assertion: %v", err)
	}
	sigB64 := base64.RawURLEncoding.EncodeToString(sigBytes)

	renewReq := LoginRenewRequest{
		ID: "passkey-1",
		Response: WebAuthnAssertionResponse{
			ClientDataJSON:    renewClientDataB64,
			AuthenticatorData: authDataB64,
			Signature:         sigB64,
		},
	}

	renewedRT, renewedAT, err := authSvc.LoginRenew(passkey, renewReq, newRT, "127.0.0.1", "Go-Test-Agent/1.0")
	if err != nil {
		t.Fatalf("LoginRenew failed: %v", err)
	}
	if renewedRT == "" || renewedRT == newRT {
		t.Fatalf("expected new renewed refresh token, got: %s", renewedRT)
	}
	if renewedAT == "" {
		t.Fatalf("expected non-empty renewed access token")
	}

	// Verify old passkey is invalidated
	var oldPasskey domain.AuthPasskey
	db.Table("T_Auth_Passkey").Where("PasskeyId = ?", passkey.PasskeyID).First(&oldPasskey)
	if oldPasskey.Status == nil || *oldPasskey.Status != false {
		t.Fatalf("expected old passkey status to be false (invalidated), got %v", oldPasskey.Status)
	}

	// Verify new passkey is saved with status true and same CredentialID
	var latestPasskey domain.AuthPasskey
	err = db.Table("T_Auth_Passkey").Where("RefreshToken = ?", renewedRT).First(&latestPasskey).Error
	if err != nil || latestPasskey.Status == nil || !*latestPasskey.Status || latestPasskey.Credential != passkey.Credential || latestPasskey.CredentialID != passkey.CredentialID {
		t.Fatalf("expected active new passkey with same credential and CredentialID, got: %+v, err: %v", latestPasskey, err)
	}

	// 4. Test GetLoginOptions
	t.Run("GetLoginOptions_SinglePasskey", func(t *testing.T) {
		opts, err := authSvc.GetLoginOptions("admin", "https://account.docusign.com")
		if err != nil {
			t.Fatalf("GetLoginOptions failed: %v", err)
		}
		if opts.Challenge == "" {
			t.Errorf("expected non-empty challenge")
		}
		if opts.RpID != "account.docusign.com" {
			t.Errorf("expected rpId 'account.docusign.com', got '%s'", opts.RpID)
		}
		if opts.UserVerification != "required" {
			t.Errorf("expected userVerification 'required', got '%s'", opts.UserVerification)
		}
		if len(opts.AllowCredentials) != 1 {
			t.Fatalf("expected 1 allowed credential, got %d", len(opts.AllowCredentials))
		}
		if opts.AllowCredentials[0].Type != "public-key" {
			t.Errorf("expected type 'public-key', got '%s'", opts.AllowCredentials[0].Type)
		}
		if opts.AllowCredentials[0].ID == "" {
			t.Errorf("expected non-empty ID")
		}
		if len(opts.AllowCredentials[0].Transports) == 0 {
			t.Errorf("expected non-empty transports")
		}
	})

	t.Run("GetLoginOptions_MultiplePasskeys_1ToMany", func(t *testing.T) {
		// Insert another passkey for adminID
		st := true
		now := time.Now()
		rt := "second-refresh-token-xyz"
		db.Table("T_Auth_Passkey").Create(&domain.AuthPasskey{
			LoginID:      adminID,
			CredentialID: "second-device-cred-id",
			Credential:   passkey.Credential,
			RefreshToken: &rt,
			CreatedAt:    &now,
			Status:       &st,
		})

		opts, err := authSvc.GetLoginOptions("admin", "http://localhost:8080")
		if err != nil {
			t.Fatalf("GetLoginOptions with multiple passkeys failed: %v", err)
		}
		if len(opts.AllowCredentials) != 2 {
			t.Fatalf("expected 2 credentials for 1-to-Many user passkeys, got %d", len(opts.AllowCredentials))
		}
	})

	t.Run("GetLoginOptions_IgnoreNullCredentialId", func(t *testing.T) {
		// Passkey with NULL CredentialId should NOT be included in GetLoginOptions
		st := true
		now := time.Now()
		rt := "null-cred-refresh-token"
		db.Table("T_Auth_Passkey").Create(&domain.AuthPasskey{
			LoginID:      adminID,
			CredentialID: "", // empty/null
			Credential:   passkey.Credential,
			RefreshToken: &rt,
			CreatedAt:    &now,
			Status:       &st,
		})

		opts, err := authSvc.GetLoginOptions("admin", "http://localhost:8080")
		if err != nil {
			t.Fatalf("GetLoginOptions failed: %v", err)
		}
		for _, cred := range opts.AllowCredentials {
			if cred.ID == "" {
				t.Fatalf("expected no empty credential ID in allowCredentials")
			}
		}
	})

	t.Run("GetLoginOptions_UserNotFound", func(t *testing.T) {
		_, err := authSvc.GetLoginOptions("unknown_user", "http://localhost:8080")
		if err == nil || err.Error() != "user not found" {
			t.Errorf("expected 'user not found' error, got %v", err)
		}
	})

	t.Run("GetLoginOptions_EmptyUsername", func(t *testing.T) {
		_, err := authSvc.GetLoginOptions("", "http://localhost:8080")
		if err == nil || err.Error() != "username is required" {
			t.Errorf("expected 'username is required' error, got %v", err)
		}
	})

	// 5. Test LoginPasskey
	t.Run("LoginPasskey_WithoutRememberMe", func(t *testing.T) {
		opts, err := authSvc.GetLoginOptions("admin", "https://app.local")
		if err != nil {
			t.Fatalf("GetLoginOptions failed: %v", err)
		}

		clientDataJSON := `{"type":"webauthn.get","challenge":"` + opts.Challenge + `","origin":"https://app.local"}`
		clientDataB64 := base64.RawURLEncoding.EncodeToString([]byte(clientDataJSON))

		authData := make([]byte, 37)
		rpIDHash := sha256.Sum256([]byte("app.local"))
		copy(authData[0:32], rpIDHash[:])
		authData[32] = 0x01 // UP
		authDataB64 := base64.RawURLEncoding.EncodeToString(authData)

		clientDataHash := sha256.Sum256([]byte(clientDataJSON))
		signedData := append(authData, clientDataHash[:]...)
		signedHash := sha256.Sum256(signedData)

		sigBytes, err := ecdsa.SignASN1(rand.Reader, ecKey, signedHash[:])
		if err != nil {
			t.Fatalf("failed to sign assertion: %v", err)
		}
		sigB64 := base64.RawURLEncoding.EncodeToString(sigBytes)

		passkeyReq := LoginPasskeyRequest{
			Username: "admin",
			Credential: LoginPasskeyCredential{
				ID: "passkey-1",
				Response: WebAuthnAssertionResponse{
					ClientDataJSON:    clientDataB64,
					AuthenticatorData: authDataB64,
					Signature:         sigB64,
				},
			},
		}

		resp, err := authSvc.LoginPasskey(passkeyReq, "192.168.1.100", "Mozilla/5.0")
		if err != nil {
			t.Fatalf("LoginPasskey without rememberme failed: %v", err)
		}
		if resp.Token == "" {
			t.Fatalf("expected non-empty access token")
		}
		if resp.User.Username != "admin" {
			t.Errorf("expected user admin, got %s", resp.User.Username)
		}
		if resp.RefreshToken != nil {
			t.Errorf("expected nil refresh_token when rememberme is false")
		}
	})

	t.Run("LoginPasskey_WithRememberMe_IPv6", func(t *testing.T) {
		opts, err := authSvc.GetLoginOptions("admin", "https://app.local")
		if err != nil {
			t.Fatalf("GetLoginOptions failed: %v", err)
		}

		clientDataJSON := `{"type":"webauthn.get","challenge":"` + opts.Challenge + `","origin":"https://app.local"}`
		clientDataB64 := base64.RawURLEncoding.EncodeToString([]byte(clientDataJSON))

		authData := make([]byte, 37)
		rpIDHash := sha256.Sum256([]byte("app.local"))
		copy(authData[0:32], rpIDHash[:])
		authData[32] = 0x01
		authDataB64 := base64.RawURLEncoding.EncodeToString(authData)

		clientDataHash := sha256.Sum256([]byte(clientDataJSON))
		signedData := append(authData, clientDataHash[:]...)
		signedHash := sha256.Sum256(signedData)

		sigBytes, err := ecdsa.SignASN1(rand.Reader, ecKey, signedHash[:])
		if err != nil {
			t.Fatalf("failed to sign assertion: %v", err)
		}
		sigB64 := base64.RawURLEncoding.EncodeToString(sigBytes)

		rememberTrue := true
		passkeyReq := LoginPasskeyRequest{
			Username: "admin",
			Credential: LoginPasskeyCredential{
				ID: "passkey-1",
				Response: WebAuthnAssertionResponse{
					ClientDataJSON:    clientDataB64,
					AuthenticatorData: authDataB64,
					Signature:         sigB64,
				},
			},
			RememberMe: &rememberTrue,
		}

		resp, err := authSvc.LoginPasskey(passkeyReq, "2001:0db8:85a3:0000:0000:8a2e:0370:7334", "Mozilla/5.0-Mobile")
		if err != nil {
			t.Fatalf("LoginPasskey with rememberme failed: %v", err)
		}
		if resp.Token == "" {
			t.Fatalf("expected non-empty access token")
		}
		if resp.RefreshToken == nil || *resp.RefreshToken == "" {
			t.Fatalf("expected non-empty refresh_token when rememberme is true")
		}
		if resp.PublicKey == nil || *resp.PublicKey == "" {
			t.Fatalf("expected non-empty public_key when rememberme is true")
		}

		// Verify record exists in DB with IPv6 and UserAgent
		var sessionRecord domain.AuthPasskey
		err = db.Table("T_Auth_Passkey").Where("RefreshToken = ?", *resp.RefreshToken).First(&sessionRecord).Error
		if err != nil {
			t.Fatalf("failed to find new rememberme passkey record in DB: %v", err)
		}
		if sessionRecord.IP == nil || *sessionRecord.IP != "2001:0db8:85a3:0000:0000:8a2e:0370:7334" {
			t.Errorf("expected IPv6 stored in DB, got %v", sessionRecord.IP)
		}
		if sessionRecord.UserAgent == nil || *sessionRecord.UserAgent != "Mozilla/5.0-Mobile" {
			t.Errorf("expected UserAgent stored in DB, got %v", sessionRecord.UserAgent)
		}
	})
}
