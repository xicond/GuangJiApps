package domain_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"testing"

	"guangjiapps/gin/internal/domain"
)

func TestFlexString_Scan(t *testing.T) {
	var fs domain.FlexString

	var scanner sql.Scanner = &fs
	if err := scanner.Scan(int64(34806)); err != nil {
		t.Fatalf("failed to scan int64: %v", err)
	}
	if string(fs) != "34806" {
		t.Fatalf("expected '34806', got '%s'", fs)
	}

	if err := scanner.Scan(int32(35649)); err != nil {
		t.Fatalf("failed to scan int32: %v", err)
	}
	if string(fs) != "35649" {
		t.Fatalf("expected '35649', got '%s'", fs)
	}

	if err := scanner.Scan(" 25440003 "); err != nil {
		t.Fatalf("failed to scan string: %v", err)
	}
	if string(fs) != "25440003" {
		t.Fatalf("expected '25440003', got '%s'", fs)
	}

	if err := scanner.Scan([]byte("99999")); err != nil {
		t.Fatalf("failed to scan []byte: %v", err)
	}
	if string(fs) != "99999" {
		t.Fatalf("expected '99999', got '%s'", fs)
	}

	if err := scanner.Scan(nil); err != nil {
		t.Fatalf("failed to scan nil: %v", err)
	}
	if string(fs) != "" {
		t.Fatalf("expected empty string, got '%s'", fs)
	}
}

func TestFlexString_Value(t *testing.T) {
	fs1 := domain.FlexString("34806")
	val1, err := fs1.Value()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val1 != driver.Value("34806") {
		t.Fatalf("expected '34806', got %v", val1)
	}

	fsEmpty := domain.FlexString("")
	valEmpty, err := fsEmpty.Value()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valEmpty != nil {
		t.Fatalf("expected nil, got %v", valEmpty)
	}
}

func TestFlexString_JSON(t *testing.T) {
	// Test unmarshal from integer
	var fs domain.FlexString
	if err := json.Unmarshal([]byte("34806"), &fs); err != nil {
		t.Fatalf("failed to unmarshal int: %v", err)
	}
	if string(fs) != "34806" {
		t.Fatalf("expected '34806', got '%s'", fs)
	}

	// Test unmarshal from string
	if err := json.Unmarshal([]byte(`"35649"`), &fs); err != nil {
		t.Fatalf("failed to unmarshal string: %v", err)
	}
	if string(fs) != "35649" {
		t.Fatalf("expected '35649', got '%s'", fs)
	}

	// Test marshal
	b, err := json.Marshal(fs)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	if string(b) != `"35649"` {
		t.Fatalf("expected '\"35649\"', got %s", string(b))
	}

	// Test marshal empty
	fsEmpty := domain.FlexString("")
	bEmpty, err := json.Marshal(fsEmpty)
	if err != nil {
		t.Fatalf("failed to marshal empty: %v", err)
	}
	if string(bEmpty) != "null" {
		t.Fatalf("expected 'null', got %s", string(bEmpty))
	}
}
