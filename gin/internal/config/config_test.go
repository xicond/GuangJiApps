package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

func TestPEMKeyLoading(t *testing.T) {
	// 1. Generate RSA Private Key for testing
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Create temp directory for PEM file
	tempDir := t.TempDir()
	pemFilePath := filepath.Join(tempDir, "jwt-test.pem")
	if err := os.WriteFile(pemFilePath, privPEM, 0600); err != nil {
		t.Fatalf("failed to write PEM file: %v", err)
	}

	cfg := Config{
		JWTSecret: pemFilePath,
	}

	// 2. Test GetJWTSigningKey
	signKey, signMethod, err := cfg.GetJWTSigningKey()
	if err != nil {
		t.Fatalf("GetJWTSigningKey failed: %v", err)
	}
	if signMethod.Alg() != "RS256" {
		t.Errorf("expected signing method RS256, got %s", signMethod.Alg())
	}
	if _, ok := signKey.(*rsa.PrivateKey); !ok {
		t.Errorf("expected *rsa.PrivateKey type for signing key")
	}

	// 3. Test GetJWTVerificationKey
	verifyKey, verifyMethod, err := cfg.GetJWTVerificationKey()
	if err != nil {
		t.Fatalf("GetJWTVerificationKey failed: %v", err)
	}
	if verifyMethod.Alg() != "RS256" {
		t.Errorf("expected verification method RS256, got %s", verifyMethod.Alg())
	}
	if _, ok := verifyKey.(*rsa.PublicKey); !ok {
		t.Errorf("expected *rsa.PublicKey type for verification key")
	}

	// 4. Test Raw Secret fallback
	rawCfg := Config{
		JWTSecret: "my-plain-secret-key-12345",
	}
	rawKey, rawMethod, err := rawCfg.GetJWTSigningKey()
	if err != nil {
		t.Fatalf("raw GetJWTSigningKey failed: %v", err)
	}
	if rawMethod.Alg() != "HS256" {
		t.Errorf("expected signing method HS256 for raw secret, got %s", rawMethod.Alg())
	}
	if string(rawKey.([]byte)) != "my-plain-secret-key-12345" {
		t.Errorf("expected raw secret bytes to match input string")
	}
}
