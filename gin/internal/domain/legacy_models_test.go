package domain

import (
	"encoding/json"
	"testing"
)

func TestIntBoolScan(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected IntBool
		wantErr  bool
	}{
		{"Nil input", nil, false, false},
		{"Bool true", true, true, false},
		{"Bool false", false, false, false},
		{"Int 1", int(1), true, false},
		{"Int 0", int(0), false, false},
		{"Int32 1", int32(1), true, false},
		{"Int32 0", int32(0), false, false},
		{"Int64 1", int64(1), true, false},
		{"Int64 0", int64(0), false, false},
		{"Int16 1", int16(1), true, false},
		{"Int8 1", int8(1), true, false},
		{"Uint8 1 (bit/tinyint)", uint8(1), true, false},
		{"Uint8 0", uint8(0), false, false},
		{"Float64 1.0", float64(1.0), true, false},
		{"Float64 0.0", float64(0.0), false, false},
		{"String '1'", "1", true, false},
		{"String '0'", "0", false, false},
		{"String 'true'", "true", true, false},
		{"String 'false'", "false", false, false},
		{"Bytes '1'", []byte("1"), true, false},
		{"Bytes '0'", []byte("0"), false, false},
		{"Invalid type struct", struct{}{}, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ib IntBool
			err := ib.Scan(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("IntBool.Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && ib != tt.expected {
				t.Errorf("IntBool.Scan() = %v, expected %v", ib, tt.expected)
			}
		})
	}
}

func TestIntBoolJSON(t *testing.T) {
	type Sample struct {
		Flag IntBool `json:"flag"`
	}

	// Test Unmarshal
	unmarshalTests := []struct {
		jsonStr  string
		expected bool
	}{
		{`{"flag": true}`, true},
		{`{"flag": false}`, false},
		{`{"flag": 1}`, true},
		{`{"flag": 0}`, false},
		{`{"flag": "1"}`, true},
		{`{"flag": "0"}`, false},
		{`{"flag": "true"}`, true},
		{`{"flag": "false"}`, false},
		{`{"flag": null}`, false},
	}

	for _, tt := range unmarshalTests {
		var s Sample
		if err := json.Unmarshal([]byte(tt.jsonStr), &s); err != nil {
			t.Errorf("Unmarshal failed for %s: %v", tt.jsonStr, err)
		} else if bool(s.Flag) != tt.expected {
			t.Errorf("Unmarshal %s = %v, expected %v", tt.jsonStr, s.Flag, tt.expected)
		}
	}

	// Test Marshal
	sTrue := Sample{Flag: true}
	bTrue, err := json.Marshal(sTrue)
	if err != nil || string(bTrue) != `{"flag":true}` {
		t.Errorf("Marshal true failed, got: %s, err: %v", string(bTrue), err)
	}

	sFalse := Sample{Flag: false}
	bFalse, err := json.Marshal(sFalse)
	if err != nil || string(bFalse) != `{"flag":false}` {
		t.Errorf("Marshal false failed, got: %s, err: %v", string(bFalse), err)
	}
}

func TestGetFieldLabel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Umat.NamaIndonesia", "Nama Indonesia"},
		{"Umat.NamaMandarin", "Nama Mandarin"},
		{"Umat.TanggalChiutaoInt", "Tanggal Ciu Tao (Masehi)"},
		{"Umat.Kode", "Kode"},
		{"Umat.NamaLengkap", "Nama Lengkap"},
		{"NamaIndonesia", "Nama Indonesia"},
		{"TempatLahir", "Tempat Lahir"},
		{"EventCode", "Event Code"},
		{"Username", "Username"},
		{"nama_indonesia", "Nama Indonesia"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := GetFieldLabel(tt.input)
			if got != tt.expected {
				t.Errorf("GetFieldLabel(%q) = %q, expected %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestDateTimeJSON(t *testing.T) {
	type Sample struct {
		Created DateTime `json:"created"`
	}

	// Test Unmarshal ISO/RFC3339 string (e.g. 2022-10-08T09:12:39.423Z)
	jsonStr := `{"created":"2022-10-08T09:12:39.423Z"}`
	var s Sample
	if err := json.Unmarshal([]byte(jsonStr), &s); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	expectedYear, expectedMonth, expectedDay := 2022, 10, 8
	expectedHour, expectedMin, expectedSec := 9, 12, 39
	if s.Created.Year() != expectedYear || int(s.Created.Month()) != expectedMonth || s.Created.Day() != expectedDay ||
		s.Created.Hour() != expectedHour || s.Created.Minute() != expectedMin || s.Created.Second() != expectedSec {
		t.Errorf("Unmarshal got %v, expected 2022-10-08 09:12:39", s.Created)
	}

	// Test Marshal outputs YYYY-MM-DD HH:mm:ss format
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	expectedJSON := `{"created":"2022-10-08 09:12:39"}`
	if string(b) != expectedJSON {
		t.Errorf("Marshal got %s, expected %s", string(b), expectedJSON)
	}
}

func TestDateTimeScan(t *testing.T) {
	var dt DateTime
	if err := dt.Scan("2022-10-08 09:12:39"); err != nil {
		t.Errorf("Scan string failed: %v", err)
	}
	if dt.Format("2006-01-02 15:04:05") != "2022-10-08 09:12:39" {
		t.Errorf("Scan string got %s", dt.Format("2006-01-02 15:04:05"))
	}
}


