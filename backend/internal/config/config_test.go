package config

import (
	"strings"
	"testing"
)

func TestLoadValidation(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr string
	}{
		{
			name: "development accepts placeholder secret",
			env:  map[string]string{"APP_ENV": "development", "JWT_SECRET": "please-change-me-to-a-long-random-string"},
		},
		{
			name:    "production rejects placeholder secret",
			env:     map[string]string{"APP_ENV": "production", "JWT_SECRET": "please-change-me-to-a-long-random-string"},
			wantErr: "JWT_SECRET",
		},
		{
			name: "production accepts strong secret",
			env: map[string]string{
				"APP_ENV":              "production",
				"JWT_SECRET":           "8f0e5f7f5b4d4c1aa2c0d5b8e3d7a4f9c6b2e1d0a9f8c7b6e5d4c3b2a1f0e9d8c",
				"CORS_ALLOWED_ORIGINS": "https://gigmatch.example.com",
			},
		},
		{
			name: "production rejects wildcard cors",
			env: map[string]string{
				"APP_ENV":              "production",
				"JWT_SECRET":           "8f0e5f7f5b4d4c1aa2c0d5b8e3d7a4f9c6b2e1d0a9f8c7b6e5d4c3b2a1f0e9d8c",
				"CORS_ALLOWED_ORIGINS": "*",
			},
			wantErr: "CORS_ALLOWED_ORIGINS",
		},
		{
			name:    "invalid server port",
			env:     map[string]string{"APP_ENV": "development", "SERVER_PORT": "70000"},
			wantErr: "SERVER_PORT",
		},
		{
			name:    "invalid auth rate limit",
			env:     map[string]string{"APP_ENV": "development", "AUTH_RATE_LIMIT": "0"},
			wantErr: "AUTH_RATE_LIMIT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, key := range []string{
				"APP_ENV", "SERVER_PORT", "JWT_SECRET", "JWT_EXPIRE_HOURS", "UPLOAD_DIR",
				"CORS_ALLOWED_ORIGINS", "AUTH_RATE_LIMIT", "API_RATE_LIMIT",
				"DB_MAX_OPEN_CONNS", "DB_MAX_IDLE_CONNS", "DB_CONN_MAX_LIFETIME_MIN",
				"DB_CONNECT_RETRIES", "DB_CONNECT_RETRY_INTERVAL_SEC",
			} {
				t.Setenv(key, "")
			}
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			cfg, err := Load()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Load() unexpected error: %v", err)
				}
				if cfg == nil {
					t.Fatal("Load() returned nil config")
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Load() error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestAllowedOrigins(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{name: "empty", input: "", want: nil},
		{name: "single", input: "http://localhost:28030", want: []string{"http://localhost:28030"}},
		{name: "multiple", input: " http://a.example , https://b.example ", want: []string{"http://a.example", "https://b.example"}},
		{name: "wildcard disabled", input: "*", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := (&Config{CORSAllowedOrigins: tt.input}).AllowedOrigins()
			if len(got) != len(tt.want) {
				t.Fatalf("AllowedOrigins() = %#v, want %#v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("AllowedOrigins() = %#v, want %#v", got, tt.want)
				}
			}
		})
	}
}
