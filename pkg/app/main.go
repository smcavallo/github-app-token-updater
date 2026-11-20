package main

import (
	"errors"
	"log/slog"
	"os"
	"strings"
)

const (
	envClientID = "GITHUB_APP_CLIENT_ID"
	envAppID    = "GITHUB_APP_ID"

	issuerTypeClientID = "client_id"
	issuerTypeAppID    = "app_id"
)

// resolveIssuer returns the value to use for the JWT `iss` claim and which kind of
// identifier it is. GitHub recommends the app's client ID, so GITHUB_APP_CLIENT_ID is
// preferred. GITHUB_APP_ID is still accepted by GitHub and is used as a fallback when the
// client ID is unset (or blank).
// https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/generating-a-json-web-token-jwt-for-a-github-app
func resolveIssuer(getenv func(string) string) (issuer string, issuerType string, err error) {
	if clientID := strings.TrimSpace(getenv(envClientID)); clientID != "" {
		return clientID, issuerTypeClientID, nil
	}
	if appID := strings.TrimSpace(getenv(envAppID)); appID != "" {
		return appID, issuerTypeAppID, nil
	}
	return "", "", errors.New("neither " + envClientID + " nor " + envAppID + " is set")
}

// Main function
func main() {
	// Log as key=value text to stderr
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	// Get the new token from GitHub
	slog.Info("Starting main function")
	issuer, issuerType, issuerErr := resolveIssuer(os.Getenv)
	if issuerErr != nil {
		slog.Error("Error resolving github app identifier", "error", issuerErr)
		os.Exit(1)
	}
	installationID := os.Getenv("GITHUB_INSTALLATION_ID")
	pemPath := "/tmp/private-key.pem"
	slog.Info("Refreshing token", "issuer", issuer, "issuer_type", issuerType, "installation_id", installationID)
	jwt, jwtErr := GenerateJWT(issuer, pemPath)
	if jwtErr != nil {
		slog.Error("Error generating jwt", "error", jwtErr)
		os.Exit(1)
	}
	token, tokenErr := GetGithubAppInstallationToken(jwt, installationID)
	if tokenErr != nil {
		slog.Error("Error generating github app token", "error", tokenErr)
		os.Exit(1)
	}
	// Update the kubernetes secret
	slog.Info("Updating kubernetes secret with new token")
	namespace := os.Getenv("SECRET_NAMESPACE")
	secretName := os.Getenv("SECRET_NAME")
	secretKey := os.Getenv("SECRET_KEY")
	UpdateKubernetesSecret(namespace, secretName, secretKey, token)
	slog.Info("Ending main function")
}
