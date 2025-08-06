package filebrowser

import (
	"fmt"
	"log"
	"path/filepath"
)

// ActionParams contains parameters for file operations
type ActionParams struct {
	ShareParams ShareParams
	FileSize    int64
	Force       bool
}

// ShareParams contains parameters for sharing files
type ShareParams struct {
	Expires  int64  // Expiration time
	Password string // Optional password protection
	Unit     string // Time unit (e.g., "hours", "days")
}

// ShareResult contains the URLs for viewing and downloading shared files
type ShareResult struct {
	ViewUrl     string
	DownloadUrl string
}

// FilebrowserAuth contains authentication credentials for Filebrowser
type FilebrowserAuth struct {
	URL      string
	Username string
	Password string
}

// Validate checks if the authentication credentials are valid
func (auth *FilebrowserAuth) Validate() error {
	if auth.URL == "" {
		return fmt.Errorf("URL cannot be empty")
	}
	if auth.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if auth.Password == "" {
		return fmt.Errorf("password cannot be empty")
	}
	return nil
}

// SaveAndShare downloads a file from an external URL, uploads it to Filebrowser,
// and creates a share link. It handles file size comparison and force overwrite.
func SaveAndShare(auth FilebrowserAuth, externalURL string, remotePathFn func(string) string, actionParams ActionParams) (*ShareResult, error) {
	// Validate authentication
	if err := auth.Validate(); err != nil {
		return nil, fmt.Errorf("invalid authentication: %w", err)
	}

	// Validate input parameters
	if externalURL == "" {
		return nil, fmt.Errorf("external URL cannot be empty")
	}
	if remotePathFn == nil {
		return nil, fmt.Errorf("remote path function cannot be nil")
	}

	// Download file to local
	localPath, err := DownloadToLocal(externalURL, actionParams.FileSize)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	// Generate remote path
	name := filepath.Base(localPath)
	remotePath := remotePathFn(name)
	if remotePath == "" {
		return nil, fmt.Errorf("remote path cannot be empty")
	}

	// Create client and authenticate
	client := &Client{
		URL: auth.URL,
		ReqLogin: ReqLogin{
			Username: auth.Username,
			Password: auth.Password,
		},
	}

	// Check if resource exists and handle size comparison
	resourceRet, err := client.GetResource(remotePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource info: %w", err)
	}

	// Handle file size comparison and force overwrite
	shouldUpload := true
	if !resourceRet.NotExist {
		if actionParams.Force {
			log.Printf("Force flag set, deleting existing resource: %s", remotePath)
			if err := client.DeleteResource(remotePath); err != nil {
				return nil, fmt.Errorf("failed to delete existing resource: %w", err)
			}
		} else if actionParams.FileSize > 0 && resourceRet.Size != actionParams.FileSize {
			log.Printf("File size mismatch, deleting existing resource: %s (local: %d, remote: %d)", 
				remotePath, actionParams.FileSize, resourceRet.Size)
			if err := client.DeleteResource(remotePath); err != nil {
				return nil, fmt.Errorf("failed to delete mismatched resource: %w", err)
			}
		} else {
			log.Printf("Resource already exists with same size, skipping upload: %s", remotePath)
			shouldUpload = false
		}
	}

	// Upload file if needed
	if shouldUpload {
		if err := client.Upload(localPath, remotePath); err != nil {
			return nil, fmt.Errorf("failed to upload file: %w", err)
		}
		log.Printf("Successfully uploaded file to: %s", remotePath)
	}

	// Create share
	hash, err := client.Share(remotePath, actionParams.ShareParams.Expires, 
		actionParams.ShareParams.Password, actionParams.ShareParams.Unit)
	if err != nil {
		return nil, fmt.Errorf("failed to create share: %w", err)
	}

	result := &ShareResult{
		ViewUrl:     fmt.Sprintf("%s/share/%s", client.URL, hash),
		DownloadUrl: fmt.Sprintf("%s/api/public/dl/%s", client.URL, hash),
	}

	log.Printf("Successfully created share: %s", result.ViewUrl)
	return result, nil
}
