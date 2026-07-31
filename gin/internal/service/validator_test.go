package service

import (
	"strings"
	"testing"

	"guangjiapps/gin/internal/domain"
)

func TestValidateStruct(t *testing.T) {
	t.Run("validation failure on missing required fields", func(t *testing.T) {
		payload := domain.Admin{}
		err := ValidateStruct(payload)
		if err == nil {
			t.Fatalf("expected validation error, got nil")
		}
		if !strings.Contains(err.Error(), "wajib diisi") {
			t.Errorf("expected error message to contain 'wajib diisi', got: %v", err)
		}
		if !strings.Contains(strings.ToLower(err.Error()), "username") {
			t.Errorf("expected error message to mention 'username', got: %v", err)
		}
	})

	t.Run("validation success when required fields provided", func(t *testing.T) {
		payload := domain.Admin{
			Username: "testuser",
		}
		err := ValidateStruct(payload)
		if err != nil {
			t.Errorf("expected no validation error, got: %v", err)
		}
	})

	t.Run("validation failure for Kelas missing required fields", func(t *testing.T) {
		payload := domain.Kelas{}
		err := ValidateStruct(payload)
		if err == nil {
			t.Fatalf("expected validation error for empty Kelas, got nil")
		}
		if !strings.Contains(strings.ToLower(err.Error()), "kode_kelas") {
			t.Errorf("expected error message to mention 'kode_kelas', got: %v", err)
		}
	})

	t.Run("validation failure when string exceeds max length", func(t *testing.T) {
		payload := domain.Admin{
			Username: strings.Repeat("a", 51),
		}
		err := ValidateStruct(payload)
		if err == nil {
			t.Fatalf("expected max length validation error, got nil")
		}
		if !strings.Contains(err.Error(), "max:50") {
			t.Errorf("expected error message to contain 'max:50', got: %v", err)
		}
	})

	t.Run("validation failure when EventCode exceeds max length of 10", func(t *testing.T) {
		payload := domain.Activity{
			EventCode: "EVENT_CODE_TOO_LONG",
			EventName: "Valid Event",
		}
		err := ValidateStruct(payload)
		if err == nil {
			t.Fatalf("expected max length validation error, got nil")
		}
		if !strings.Contains(err.Error(), "max:10") {
			t.Errorf("expected error message to contain 'max:10', got: %v", err)
		}
	})

	t.Run("validation failure on invalid email format", func(t *testing.T) {
		emailStr := "invalid-email-format"
		payload := domain.Admin{
			Username: "admin_user",
			Email:    &emailStr,
		}
		err := ValidateStruct(payload)
		if err == nil {
			t.Fatalf("expected email validation error, got nil")
		}
		if !strings.Contains(err.Error(), "email") {
			t.Errorf("expected error message to contain 'email', got: %v", err)
		}
	})

	t.Run("validation failure on negative amount gte=0", func(t *testing.T) {
		payload := domain.DonasiSxy{
			NoKwitansi: "KW001",
			Jumlah:     -500.0,
		}
		err := ValidateStruct(payload)
		if err == nil {
			t.Fatalf("expected gte validation error, got nil")
		}
		if !strings.Contains(err.Error(), "kurang dari nilai minimum 0") {
			t.Errorf("expected error message to contain 'kurang dari nilai minimum 0', got: %v", err)
		}
	})

	t.Run("DateOnly JSON unmarshaling", func(t *testing.T) {
		var d domain.DateOnly
		err := d.UnmarshalJSON([]byte(`"2026-07-27"`))
		if err != nil {
			t.Errorf("failed to unmarshal DateOnly: %v", err)
		}
		if d.Format("2006-01-02") != "2026-07-27" {
			t.Errorf("expected 2026-07-27, got %s", d.Format("2006-01-02"))
		}

		err = d.UnmarshalJSON([]byte(`"invalid-date"`))
		if err == nil {
			t.Errorf("expected error for invalid date, got nil")
		}
	})

	t.Run("validation failure translates field name using dictionary mapping or space separation", func(t *testing.T) {
		payload := domain.Umat{}
		err := ValidateStruct(payload)
		if err == nil {
			t.Fatalf("expected validation error for empty Umat, got nil")
		}
		if !strings.Contains(err.Error(), "Nama Indonesia wajib diisi") {
			t.Errorf("expected translated error message for NamaIndonesia, got: %v", err)
		}
	})
}

