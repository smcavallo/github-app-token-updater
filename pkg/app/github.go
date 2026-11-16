package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

const (
	errGetJwtToken      = "error creating token: %w"
	errPemDoesNotExist  = "error private-key.pem does not exist: %w"
	errPemRead          = "error reading private-key.pem: %w"
	errPemParse         = "error parsing private-key.pem as an RSA private key: %w"
	errHttpRequestError = "error making http request: %w"
	errJSONUnmarshal    = "unable to unmarshal token response: %w"
)

// GenerateJWT calls the GitHub api to get a JWT
// https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/generating-a-json-web-token-jwt-for-a-github-app#example-using-python-to-generate-a-jwt
func GenerateJWT(appID string, pemFilePath string) (string, error) {
	now := time.Now()

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": jwt.NewNumericDate(now.Add(-time.Minute)),
		"exp": jwt.NewNumericDate(now.Add(5 * time.Minute)),
		"iss": appID,
	})

	pemKey, err := os.ReadFile(pemFilePath)
	if errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf(errPemDoesNotExist, err)
	}
	if err != nil {
		return "", fmt.Errorf(errPemRead, err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(pemKey)
	if err != nil {
		return "", fmt.Errorf(errPemParse, err)
	}

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf(errGetJwtToken, err)
	}

	return tokenString, err
}

// GetGithubAppInstallationToken gets a token for an app installation
// https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/authenticating-as-a-github-app-installation#using-an-installation-access-token-to-authenticate-as-an-app-installation
func GetGithubAppInstallationToken(jwt string, installationId string) (string, error) {

	endpointURL := fmt.Sprintf("https://api.github.com/app/installations/%s/access_tokens", installationId)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointURL, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("error: creating github app installation token request: %w", err)
	}

	// Set Headers
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))

	// Create http client
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
	}
	client := &http.Client{Transport: tr}

	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf(errHttpRequestError, err)
	}

	// Unmarshal response
	githubAppInstallationTokenResponse := GithubAppInstallationTokenResponse{}
	err = ReadAndUnmarshal(res, &githubAppInstallationTokenResponse)
	if err != nil {
		return "", fmt.Errorf(errJSONUnmarshal, err)
	}
	logrus.Infof("Retrieved new github token with expiry %s", githubAppInstallationTokenResponse.ExpiresAt)
	return githubAppInstallationTokenResponse.Token, err
}

type GithubAppInstallationTokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"` // "2024-06-27T16:15:23Z"
}

func ReadAndUnmarshal(resp *http.Response, target any) error {
	defer func() {
		_ = resp.Body.Close()
	}()

	// Read the body before checking the status so that error responses
	// (e.g. {"message":"Bad credentials"}) are included in the returned error.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("failed to authenticate with the given credentials: %d %s", resp.StatusCode, body)
	}
	return json.Unmarshal(body, target)
}
