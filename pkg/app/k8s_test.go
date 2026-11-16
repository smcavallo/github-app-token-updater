package main

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestBuildSecretPatch(t *testing.T) {
	tests := []struct {
		name string
		key  string
		val  string
	}{
		{name: "plain key", key: "token", val: "ghs_exampletoken"},
		{name: "key with quote and backslash", key: `to"ken\x`, val: "ghs_exampletoken"},
		{name: "empty value", key: "token", val: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patch, err := buildSecretPatch(tt.key, tt.val)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var parsed struct {
				Data map[string]string `json:"data"`
			}
			if err := json.Unmarshal(patch, &parsed); err != nil {
				t.Fatalf("patch is not valid JSON: %v\n%s", err, patch)
			}
			encoded, ok := parsed.Data[tt.key]
			if !ok {
				t.Fatalf("patch does not contain data key %q: %s", tt.key, patch)
			}
			decoded, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				t.Fatalf("value is not valid base64: %v", err)
			}
			if string(decoded) != tt.val {
				t.Errorf("decoded value = %q, want %q", decoded, tt.val)
			}
		})
	}
}
