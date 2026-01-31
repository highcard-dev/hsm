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

// GetDownloadURL fetches the signed download URL for a patchline
func (s *DownloadService) GetDownloadURL(patchline string) (string, string, error) {
	resultFileInfoUrl, err := s.client.GetSignedURL(fmt.Sprintf("version/%s.json", patchline))
	if err != nil {
		return "", "", err
	}

	resp, err := http.Get(resultFileInfoUrl.URL)
	if err != nil {
		return "", "", err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	var resultFileInfo struct {
		DownloadURL string `json:"download_url"`
		Version     string `json:"version"`
	}
	if err := json.Unmarshal(body, &resultFileInfo); err != nil {
		return "", "", err
	}

	resultDownloadURL, err := s.client.GetSignedURL(resultFileInfo.DownloadURL)
	if err != nil {
		return "", "", err
	}
	return resultDownloadURL.URL, resultFileInfo.Version, nil
}
