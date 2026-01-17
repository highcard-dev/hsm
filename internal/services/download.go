package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"hsm/internal/client"
)

const (
	PatchlineRelease    = "release"
	PatchlinePrerelease = "prerelease"

	downloadBaseURLRelease = "https://downloader.hytale.com"
)

type DownloadService struct {
	httpClient *http.Client
	client     *client.Client
}

func NewDownloadService(c *client.Client) *DownloadService {
	return &DownloadService{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		client:     c,
	}
}

// fetchJSON makes a GET request and decodes the JSON response into the provided target
func (s *DownloadService) fetchJSON(url string, target any) error {
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

func (s *DownloadService) GetLatestVersion(patchline string) (string, error) {
	var result struct {
		Latest string `json:"latest"`
	}
	if err := s.fetchJSON(downloadBaseURLRelease+"/version.json", &result); err != nil {
		return "", err
	}
	return result.Latest, nil
}

// GetDownloadURL fetches the signed download URL for a patchline
func (s *DownloadService) GetDownloadURL(patchline string) (string, error) {
	result, err := s.client.GetSignedURL(patchline)
	if err != nil {
		return "", err
	}
	return result.URL, nil
}
