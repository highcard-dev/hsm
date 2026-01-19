package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// DownloadFile downloads a file from url to dest with progress bar
func DownloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	fmt.Println("Downloading...")
	pw := NewProgressWriter(resp.ContentLength)
	_, err = io.Copy(out, pw.WrapReader(resp.Body))
	pw.Finish()

	return err
}

// ExtractZip extracts a zip file to dest, showing progress
func ExtractZip(zipPath, dest string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	destAbs, err := filepath.Abs(dest)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	fmt.Printf("Extracting to %s\n", dest)
	total := len(r.File)

	for i, f := range r.File {
		printExtractProgress(i+1, total, f.Name)
		if err := extractFile(f, destAbs); err != nil {
			fmt.Println()
			return err
		}
	}

	fmt.Printf("\r  Extracted %d files%s\n", total, strings.Repeat(" ", 50))
	return nil
}

func printExtractProgress(current, total int, name string) {
	if len(name) > 50 {
		name = "..." + name[len(name)-47:]
	}
	fmt.Printf("\r  [%d/%d] %-50s", current, total, name)
}

func extractFile(f *zip.File, destAbs string) error {
	fpath := filepath.Join(destAbs, f.Name)

	// ZipSlip protection
	if !strings.HasPrefix(filepath.Clean(fpath), destAbs+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", f.Name)
	}

	if f.FileInfo().IsDir() {
		return os.MkdirAll(fpath, os.ModePerm)
	}

	if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
		return err
	}

	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, rc)
	return err
}

// DownloadAndExtract downloads a zip from url and extracts to dest
func DownloadAndExtract(url, dest string) error {
	tmp, err := os.CreateTemp("", "download-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	tmp.Close()
	defer os.Remove(tmpPath)

	if err := DownloadFile(url, tmpPath); err != nil {
		return err
	}

	fmt.Println()
	return ExtractZip(tmpPath, dest)
}
