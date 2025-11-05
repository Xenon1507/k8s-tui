package helm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// HelmTime is a custom time type that can parse Helm's time format
type HelmTime struct {
	time.Time
}

// UnmarshalJSON implements custom JSON unmarshaling for Helm's time format
func (ht *HelmTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "null" || s == "" {
		ht.Time = time.Time{}
		return nil
	}

	// Try multiple time formats that Helm might use
	formats := []string{
		"2006-01-02 15:04:05.999999999 -0700 MST", // Helm's default format
		time.RFC3339,                               // Standard RFC3339
		time.RFC3339Nano,                           // RFC3339 with nanoseconds
		"2006-01-02T15:04:05.999999999Z07:00",     // Alternative format
	}

	var err error
	for _, format := range formats {
		ht.Time, err = time.Parse(format, s)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("unable to parse time %q with any known format: %w", s, err)
}

// Release represents a Helm release
type Release struct {
	Name         string   `json:"name"`
	Namespace    string   `json:"namespace"`
	Revision     string   `json:"revision"`
	Updated      HelmTime `json:"updated"`
	Status       string   `json:"status"`
	Chart        string   `json:"chart"`
	AppVersion   string   `json:"app_version"`
	Description  string   `json:"description"`
}

// ReleaseHistory represents a single revision in release history
type ReleaseHistory struct {
	Revision    int      `json:"revision"`
	Updated     HelmTime `json:"updated"`
	Status      string   `json:"status"`
	Chart       string   `json:"chart"`
	AppVersion  string   `json:"app_version"`
	Description string   `json:"description"`
}

// Client wraps Helm CLI operations
type Client struct {
	helmPath string
}

// NewClient creates a new Helm client
func NewClient() (*Client, error) {
	// Check if helm is installed
	helmPath, err := exec.LookPath("helm")
	if err != nil {
		return nil, fmt.Errorf("helm not found in PATH: %w", err)
	}

	return &Client{
		helmPath: helmPath,
	}, nil
}

// ListReleases returns all Helm releases across all namespaces
func (c *Client) ListReleases() ([]Release, error) {
	cmd := exec.Command(c.helmPath, "list", "-A", "--output", "json")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("helm list failed: %w, stderr: %s", err, stderr.String())
	}

	var releases []Release
	if err := json.Unmarshal(stdout.Bytes(), &releases); err != nil {
		return nil, fmt.Errorf("failed to parse helm list output: %w", err)
	}

	return releases, nil
}

// GetReleaseStatus returns detailed status for a specific release
func (c *Client) GetReleaseStatus(name, namespace string) (*Release, error) {
	cmd := exec.Command(c.helmPath, "status", name, "-n", namespace, "--output", "json")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("helm status failed: %w, stderr: %s", err, stderr.String())
	}

	var result struct {
		Name      string `json:"name"`
		Info      struct {
			Status      string   `json:"status"`
			LastDeployed HelmTime `json:"last_deployed"`
			Description string   `json:"description"`
		} `json:"info"`
		Chart struct {
			Metadata struct {
				Name       string `json:"name"`
				Version    string `json:"version"`
				AppVersion string `json:"appVersion"`
			} `json:"metadata"`
		} `json:"chart"`
		Version   int    `json:"version"`
		Namespace string `json:"namespace"`
	}

	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("failed to parse helm status output: %w", err)
	}

	release := &Release{
		Name:        result.Name,
		Namespace:   result.Namespace,
		Revision:    fmt.Sprintf("%d", result.Version),
		Updated:     result.Info.LastDeployed,
		Status:      result.Info.Status,
		Chart:       fmt.Sprintf("%s-%s", result.Chart.Metadata.Name, result.Chart.Metadata.Version),
		AppVersion:  result.Chart.Metadata.AppVersion,
		Description: result.Info.Description,
	}

	return release, nil
}

// GetReleaseHistory returns the revision history for a release
func (c *Client) GetReleaseHistory(name, namespace string) ([]ReleaseHistory, error) {
	cmd := exec.Command(c.helmPath, "history", name, "-n", namespace, "--output", "json")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("helm history failed: %w, stderr: %s", err, stderr.String())
	}

	var history []ReleaseHistory
	if err := json.Unmarshal(stdout.Bytes(), &history); err != nil {
		return nil, fmt.Errorf("failed to parse helm history output: %w", err)
	}

	return history, nil
}

// GetReleaseValues returns the values for a release
func (c *Client) GetReleaseValues(name, namespace string) (string, error) {
	cmd := exec.Command(c.helmPath, "get", "values", name, "-n", namespace, "--output", "yaml")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("helm get values failed: %w, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// GetReleaseManifest returns the rendered manifest for a release
func (c *Client) GetReleaseManifest(name, namespace string) (string, error) {
	cmd := exec.Command(c.helmPath, "get", "manifest", name, "-n", namespace)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("helm get manifest failed: %w, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// IsHelmAvailable checks if Helm is installed and available
func IsHelmAvailable() bool {
	_, err := exec.LookPath("helm")
	return err == nil
}

// GetReleaseNameFromLabels extracts the Helm release name from resource labels
func GetReleaseNameFromLabels(labels map[string]string) string {
	if labels == nil {
		return ""
	}

	// Check if managed by Helm
	managedBy, ok := labels["app.kubernetes.io/managed-by"]
	if !ok || !strings.EqualFold(managedBy, "helm") {
		return ""
	}

	// Get the release instance name
	releaseName, ok := labels["app.kubernetes.io/instance"]
	if !ok {
		// Fallback to legacy label
		releaseName = labels["helm.sh/chart"]
	}

	return releaseName
}
