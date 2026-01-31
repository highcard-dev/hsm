package cmd

import (
	"fmt"
	"path/filepath"

	"hsm/internal/client"
	"hsm/internal/services"
	"hsm/internal/utils"

	"github.com/spf13/cobra"
)

var (
	print bool
)

const versionFile = "version.txt"

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the Hytale game files",
	Long:  "Update the Hytale game files to the latest version.",
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionService, err := services.NewSessionService(client.New(), GetSessionLocation())
		if err != nil {
			return fmt.Errorf("failed to initialize session: %w (run 'hsm login' first)", err)
		}
		defer sessionService.Close()
		downloadService := services.NewDownloadService(sessionService.Client())
		downloadURL, version, err := downloadService.GetDownloadURL(patchline)
		if err != nil {
			return fmt.Errorf("failed to get latest version: %w", err)
		}
		fmt.Printf("Latest Version: %s (patchline: %s)\n", version, patchline)
		if print {
			return nil
		}

		versionFile := filepath.Join(outputDir, versionFile)

		currentVersion := utils.GetVersion(versionFile)
		if currentVersion == version {
			fmt.Println("Already up to date")
			return nil
		}

		if currentVersion == "" {
			fmt.Println("No version file found, downloading latest version")
		} else {
			fmt.Println("Updating to latest version")
		}

		if err := utils.DownloadAndExtract(downloadURL, outputDir); err != nil {
			return fmt.Errorf("download failed: %w", err)
		}

		fmt.Println("Extracted to:", outputDir)

		if err := utils.WriteVersion(versionFile, version); err != nil {
			return fmt.Errorf("failed to write version: %w", err)
		}

		fmt.Println("Updated to version:", version)

		return nil
	},
}

func init() {
	updateCmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for extracted files")
	updateCmd.Flags().StringVar(&patchline, "patchline", services.PatchlineRelease, "Patchline to download (release or prerelease)")
	updateCmd.Flags().BoolVarP(&print, "print-only", "p", false, "Print the latest version")
	rootCmd.AddCommand(updateCmd)
}
