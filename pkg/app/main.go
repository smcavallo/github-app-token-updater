package main

import (
	"log/slog"
	"os"
)

// Main function
func main() {
	// Log as key=value text to stderr
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	// Get the new token from GitHub
	slog.Info("Starting main function")
	appId := os.Getenv("GITHUB_APP_ID")
	installationID := os.Getenv("GITHUB_INSTALLATION_ID")
	pemPath := "/tmp/private-key.pem"
	slog.Info("Refreshing token", "app_id", appId, "installation_id", installationID)
	jwt, jwtErr := GenerateJWT(appId, pemPath)
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
