package cmd

import (
	"fmt"
	"os"

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
		sessionService, err := services.NewSessionService(client.New(), GetSessionLocation())
		if err != nil {
			return fmt.Errorf("failed to initialize session: %w (run 'hsm login' first)", err)
		}
		defer sessionService.Close()

		downloadService := services.NewDownloadService(sessionService.Client())

		fmt.Printf("Fetching download URL for patchline: %s\n", patchline)
		url, err := downloadService.GetDownloadURL(patchline)
		if err != nil {
			return fmt.Errorf("failed to get download URL: %w", err)
		}

		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		if err := utils.DownloadAndExtract(url, outputDir); err != nil {
			return fmt.Errorf("download failed: %w", err)
		}

		fmt.Println("\nDownload and extraction complete!")
		return nil
	},
}

func init() {
	downloadCmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for extracted files")
	downloadCmd.Flags().StringVar(&patchline, "patchline", services.PatchlineRelease, "Patchline to download (release or prerelease)")
	rootCmd.AddCommand(downloadCmd)
}
