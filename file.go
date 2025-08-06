package filebrowser

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/fileutil"
	"github.com/duke-git/lancet/v2/netutil"
)

// DownloadToLocal downloads a file from the given URL to a local path.
// It checks if the file already exists with the same size to avoid re-downloading.
// Returns the local path where the file was downloaded.
func DownloadToLocal(fileURL string, fileSize int64) (string, error) {
	if fileURL == "" {
		return "", fmt.Errorf("file URL cannot be empty")
	}

	localPath := LocalPathForDownload(fileURL)
	if err := EnsureFolderForFile(localPath); err != nil {
		return "", fmt.Errorf("failed to create directory for file: %w", err)
	}

	// Check if file already exists with same size
	if fileSize > 0 && fileExistsWithSameSize(localPath, fileSize) {
		log.Printf("File already exists with same size, skipping download: %s", localPath)
		return localPath, nil
	}

	// Download the file
	if err := netutil.DownloadFile(localPath, fileURL); err != nil {
		return "", fmt.Errorf("failed to download file from %s: %w", fileURL, err)
	}

	log.Printf("Successfully downloaded file to: %s", localPath)
	return localPath, nil
}

// fileExistsWithSameSize checks if a file exists and has the same size as expected
func fileExistsWithSameSize(localPath string, expectedSize int64) bool {
	if !fileutil.IsExist(localPath) {
		return false
	}

	localSize, err := fileutil.FileSize(localPath)
	if err != nil {
		log.Printf("Warning: failed to get local file size: %v", err)
		return false
	}

	expectedSizeInt, err := convertor.ToInt(expectedSize)
	if err != nil {
		log.Printf("Warning: failed to convert expected size: %v", err)
		return false
	}

	return localSize == expectedSizeInt
}

// LocalPathForDownload generates a local path for downloading a file from a URL.
// It uses the system's temp directory as the base path.
func LocalPathForDownload(fileURL string) string {
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		log.Printf("Warning: failed to parse URL %s: %v", fileURL, err)
		// Fallback: use URL as filename
		return filepath.Join(os.TempDir(), filepath.Base(fileURL))
	}

	path := strings.TrimPrefix(parsedURL.Path, "/")
	if path == "" {
		// If path is empty, use a default filename
		path = "downloaded_file"
	}

	return filepath.Join(os.TempDir(), path)
}

// EnsureFolderForFile creates the directory structure needed for the given file path.
func EnsureFolderForFile(localPath string) error {
	if localPath == "" {
		return fmt.Errorf("local path cannot be empty")
	}

	dir := filepath.Dir(localPath)
	if err := fileutil.CreateDir(dir); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	return nil
}
