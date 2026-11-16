package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestReadAndUnmarshal_Success(t *testing.T) {
	var got GithubAppInstallationTokenResponse
	err := ReadAndUnmarshal(newResponse(http.StatusCreated, `{"token":"abc","expires_at":"2024-06-27T16:15:23Z"}`), &got)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Token != "abc" || got.ExpiresAt != "2024-06-27T16:15:23Z" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestReadAndUnmarshal_ErrorIncludesResponseBody(t *testing.T) {
	var got GithubAppInstallationTokenResponse
	err := ReadAndUnmarshal(newResponse(http.StatusUnauthorized, `{"message":"Bad credentials"}`), &got)
	if err == nil {
		t.Fatal("expected an error for a 401 response")
	}
	for _, want := range []string{"401", "Bad credentials"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not contain %q", err.Error(), want)
		}
	}
}

func TestGenerateJWT_MissingKeyFile(t *testing.T) {
	_, err := GenerateJWT("123", filepath.Join(t.TempDir(), "does-not-exist.pem"))
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected an os.ErrNotExist error, got: %v", err)
	}
}

func TestGenerateJWT_InvalidKeyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.pem")
	if err := os.WriteFile(path, []byte("this is not a pem file"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := GenerateJWT("123", path)
	if err == nil {
		t.Fatal("expected an error for an invalid PEM file")
	}
	if !strings.Contains(err.Error(), "parsing private-key.pem") {
		t.Errorf("expected a PEM parse error, got: %v", err)
	}
}

func TestGenerateJWT_ValidKey(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	path := filepath.Join(t.TempDir(), "key.pem")
	if err := os.WriteFile(path, pemBytes, 0o600); err != nil {
		t.Fatal(err)
	}

	token, err := GenerateJWT("123", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parts := strings.Split(token, "."); len(parts) != 3 {
		t.Fatalf("expected a 3-part JWT, got %d parts", len(parts))
	}
}
