package main

import "testing"

func TestResolveIssuer(t *testing.T) {
	tests := []struct {
		name       string
		env        map[string]string
		wantIssuer string
		wantType   string
		wantErr    bool
	}{
		{
			name:       "client id only",
			env:        map[string]string{"GITHUB_APP_CLIENT_ID": "Iv23liExample"},
			wantIssuer: "Iv23liExample",
			wantType:   "client_id",
		},
		{
			name:       "app id only falls back",
			env:        map[string]string{"GITHUB_APP_ID": "1234567"},
			wantIssuer: "1234567",
			wantType:   "app_id",
		},
		{
			name:       "client id wins when both are set",
			env:        map[string]string{"GITHUB_APP_CLIENT_ID": "Iv23liExample", "GITHUB_APP_ID": "1234567"},
			wantIssuer: "Iv23liExample",
			wantType:   "client_id",
		},
		{
			name:       "blank client id falls back to app id",
			env:        map[string]string{"GITHUB_APP_CLIENT_ID": "  ", "GITHUB_APP_ID": "1234567"},
			wantIssuer: "1234567",
			wantType:   "app_id",
		},
		{
			name:       "values are trimmed",
			env:        map[string]string{"GITHUB_APP_CLIENT_ID": " Iv23liExample\n"},
			wantIssuer: "Iv23liExample",
			wantType:   "client_id",
		},
		{
			name:    "neither set is an error",
			env:     map[string]string{},
			wantErr: true,
		},
		{
			name:    "both blank is an error",
			env:     map[string]string{"GITHUB_APP_CLIENT_ID": " ", "GITHUB_APP_ID": ""},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getenv := func(k string) string { return tt.env[k] }
			issuer, issuerType, err := resolveIssuer(getenv)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if issuer != tt.wantIssuer || issuerType != tt.wantType {
				t.Errorf("got (%q, %q), want (%q, %q)", issuer, issuerType, tt.wantIssuer, tt.wantType)
			}
		})
	}
}
