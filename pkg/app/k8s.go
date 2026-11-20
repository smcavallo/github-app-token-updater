package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/retry"
)

// buildSecretPatch returns a strategic merge patch that sets data[secretKey] to tokenValue.
// Marshalling with encoding/json escapes the key and base64-encodes the []byte value, so
// the patch is always valid JSON regardless of the contents of secretKey.
func buildSecretPatch(secretKey string, tokenValue string) ([]byte, error) {
	return json.Marshal(map[string]any{
		"data": map[string][]byte{secretKey: []byte(tokenValue)},
	})
}

func UpdateKubernetesSecret(namespace string, secretName string, secretKey string, tokenValue string) {
	// https://github.com/kubernetes/client-go/blob/master/examples/in-cluster-client-configuration/main.go
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	// Build the patch (the token is base64-encoded by encoding/json since it is a []byte)
	patch, err := buildSecretPatch(secretKey, tokenValue)
	if err != nil {
		panic(fmt.Errorf("building secret patch failed: %w", err))
	}

	// Attempt to patch current secret
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Update secret value
		_, updateErr := clientset.CoreV1().Secrets(namespace).Patch(context.TODO(), secretName, types.StrategicMergePatchType, patch, metav1.PatchOptions{FieldManager: "github-app-token-updater"})
		return updateErr
	})
	if retryErr != nil {
		slog.Error("failed to update secret", "secret", secretName, "namespace", namespace, "error", retryErr)
		panic(fmt.Errorf("update failed: %v", retryErr))
	}
	slog.Info("Updated kubernetes secret with new github app token", "secret", secretName, "key", secretKey, "namespace", namespace)
}
