package services

import (
	"encoding/json"
	"fmt"
	"io"
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
	resultFileInfoUrl, err := s.client.GetSignedURL(fmt.Sprintf("version/%s.json", patchline))
	if err != nil {
		return "", err
	}

	resp, err := http.Get(resultFileInfoUrl.URL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var resultFileInfo struct {
		DownloadURL string `json:"download_url"`
	}
	if err := json.Unmarshal(body, &resultFileInfo); err != nil {
		return "", err
	}

	resultDownloadURL, err := s.client.GetSignedURL(resultFileInfo.DownloadURL)
	if err != nil {
		return "", err
	}
	return resultDownloadURL.URL, nil
}
