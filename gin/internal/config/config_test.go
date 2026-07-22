package config

import "testing"

func TestNormalizeBaseURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "full url", input: "https://example.com/api", want: "https://example.com/api"},
		{name: "path only", input: "/api/v1", want: "/api/v1"},
		{name: "path without leading slash", input: "api/v1", want: "/api/v1"},
		{name: "empty", input: "", want: "/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeBaseURL(tt.input); got != tt.want {
				t.Fatalf("NormalizeBaseURL(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
