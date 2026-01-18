package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"hsm/internal/client"
	"hsm/internal/services"
	"hsm/internal/utils"

	"github.com/spf13/cobra"
)

var (
	outputDir string
	patchline string
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download and extract Hytale game files",
	Long:  "Download the Hytale game files from the specified patchline and extract them to the output directory.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load session from file
		sessionPath := GetSessionLocation()
		session, err := utils.ReadSessionFromFile(sessionPath)
		if err != nil {
			return fmt.Errorf("failed to read session from %s: %w (run 'hsm login' first)", sessionPath, err)
		}

		// Create client with token
		c := client.New().WithToken(session.Token)

		// Create download service
		downloadService := services.NewDownloadService(c)

		// Get the download URL
		fmt.Printf("Fetching download URL for patchline: %s\n", patchline)
		downloadURL, err := downloadService.GetDownloadURL(patchline)
		if err != nil {
			return fmt.Errorf("failed to get download URL: %w", err)
		}

		// Create output directory if it doesn't exist
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Download and extract
		fmt.Println("Downloading and extracting game files...")
		if err := downloadAndExtract(downloadURL, outputDir); err != nil {
			return fmt.Errorf("failed to download and extract: %w", err)
		}

		fmt.Println("Download and extraction complete!")
		return nil
	},
}

func init() {
	downloadCmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for extracted files")
	downloadCmd.Flags().StringVar(&patchline, "patchline", services.PatchlineRelease, "Patchline to download (release or prerelease)")
	rootCmd.AddCommand(downloadCmd)
}

// downloadAndExtract streams the download to a temp file and extracts from it.
// ZIP format requires random access (central directory at end of file),
// so we stream to disk to avoid holding the entire archive in memory.
func downloadAndExtract(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	// Create temp file to stream download (ZIP requires random access)
	tempFile, err := os.CreateTemp("", "hytale-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()
	defer func() { _ = os.Remove(tempPath) }()

	// Stream download to temp file (constant memory usage)
	written, err := io.Copy(tempFile, resp.Body)
	if err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("failed to download: %w", err)
	}
	_ = tempFile.Close()

	fmt.Printf("Downloaded %.2f MB\n", float64(written)/(1024*1024))

	// Open zip and extract
	fmt.Printf("Extracting to %s...\n", dest)
	zipReader, err := zip.OpenReader(tempPath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer func() { _ = zipReader.Close() }()

	for _, f := range zipReader.File {
		if err := extractZipFile(f, dest); err != nil {
			return err
		}
	}

	return nil
}

// extractZipFile extracts a single file from the zip archive
func extractZipFile(f *zip.File, dest string) error {
	fpath := filepath.Join(dest, f.Name)

	// Check for ZipSlip vulnerability
	if !strings.HasPrefix(filepath.Clean(fpath), filepath.Clean(dest)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path in zip: %s", f.Name)
	}

	if f.FileInfo().IsDir() {
		return os.MkdirAll(fpath, os.ModePerm)
	}

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Open the file in the zip
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("failed to open zip entry: %w", err)
	}
	defer func() { _ = rc.Close() }()

	// Create the destination file
	outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() { _ = outFile.Close() }()

	// Stream from zip entry to output file
	if _, err := io.Copy(outFile, rc); err != nil {
		return fmt.Errorf("failed to extract file: %w", err)
	}

	return nil
}
