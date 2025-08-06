# Filebrowser SDK

A Go SDK for interacting with Filebrowser instances. This SDK provides a clean and robust interface for downloading files, uploading them to Filebrowser, and creating share links.

## Features

- **File Download**: Download files from external URLs to local storage
- **File Upload**: Upload files to Filebrowser using TUS protocol
- **Share Management**: Create share links with optional expiration and password protection
- **Resource Management**: Check, get, and delete resources on Filebrowser
- **Robust Error Handling**: Comprehensive error handling with detailed error messages
- **Authentication**: Automatic token-based authentication with Filebrowser

## Installation

```bash
go get github.com/kiuber/filebrowser-sdk
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    "path/filepath"
    
    "filebrowser-sdk"
)

func main() {
    // Configure authentication
    auth := filebrowser.FilebrowserAuth{
        URL:      "https://your-filebrowser-instance.com",
        Username: "your-username",
        Password: "your-password",
    }

    // Define parameters
    actionParams := filebrowser.ActionParams{
        FileSize: 1024 * 1024, // 1MB
        Force:    false,
        ShareParams: filebrowser.ShareParams{
            Expires:  24,
            Password: "optional-password",
            Unit:     "hours",
        },
    }

    // Define remote path function
    remotePathFn := func(filename string) string {
        return filepath.Join("uploads", filename)
    }

    // Download, upload, and share
    result, err := filebrowser.SaveAndShare(
        auth,
        "https://example.com/file.pdf",
        remotePathFn,
        actionParams,
    )
    if err != nil {
        log.Fatalf("Failed to save and share: %v", err)
    }

    fmt.Printf("View URL: %s\n", result.ViewUrl)
    fmt.Printf("Download URL: %s\n", result.DownloadUrl)
}
```

## API Reference

### Types

#### `FilebrowserAuth`
Authentication credentials for Filebrowser.

```go
type FilebrowserAuth struct {
    URL      string
    Username string
    Password string
}
```

#### `ActionParams`
Parameters for file operations.

```go
type ActionParams struct {
    ShareParams ShareParams
    FileSize    int64
    Force       bool
}
```

#### `ShareParams`
Parameters for sharing files.

```go
type ShareParams struct {
    Expires  int64  // Expiration time
    Password string // Optional password protection
    Unit     string // Time unit (e.g., "hours", "days")
}
```

#### `ShareResult`
Result containing share URLs.

```go
type ShareResult struct {
    ViewUrl     string
    DownloadUrl string
}
```

### Functions

#### `SaveAndShare`
Downloads a file from an external URL, uploads it to Filebrowser, and creates a share link.

```go
func SaveAndShare(
    auth FilebrowserAuth,
    externalURL string,
    remotePathFn func(string) string,
    actionParams ActionParams,
) (*ShareResult, error)
```

**Parameters:**
- `auth`: Filebrowser authentication credentials
- `externalURL`: URL of the file to download
- `remotePathFn`: Function that generates the remote path from filename
- `actionParams`: Operation parameters including file size, force flag, and share settings

**Returns:**
- `*ShareResult`: Contains view and download URLs
- `error`: Any error that occurred during the operation

#### `DownloadToLocal`
Downloads a file from a URL to local storage.

```go
func DownloadToLocal(fileURL string, fileSize int64) (string, error)
```

**Parameters:**
- `fileURL`: URL of the file to download
- `fileSize`: Expected file size (0 to skip size checking)

**Returns:**
- `string`: Local path where the file was downloaded
- `error`: Any error that occurred during download

### Client Methods

#### `Client.Login()`
Authenticates with the Filebrowser server.

```go
func (c *Client) Login() error
```

#### `Client.Upload()`
Uploads a local file to Filebrowser.

```go
func (c *Client) Upload(localPath string, remotePath string) error
```

#### `Client.Share()`
Creates a share link for a file.

```go
func (c *Client) Share(remotePath string, expires int64, password string, unit string) (string, error)
```

#### `Client.GetResource()`
Retrieves information about a resource.

```go
func (c *Client) GetResource(remotePath string) (*RespResource, error)
```

#### `Client.DeleteResource()`
Deletes a resource from Filebrowser.

```go
func (c *Client) DeleteResource(remotePath string) error
```

## Error Handling

The SDK provides comprehensive error handling with detailed error messages. All functions return errors instead of panicking, allowing you to handle errors gracefully:

```go
result, err := filebrowser.SaveAndShare(auth, externalURL, remotePathFn, actionParams)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "authentication"):
        // Handle authentication errors
    case strings.Contains(err.Error(), "download"):
        // Handle download errors
    case strings.Contains(err.Error(), "upload"):
        // Handle upload errors
    default:
        // Handle other errors
    }
}
```

## Features

### File Size Comparison
The SDK automatically compares file sizes to avoid re-downloading or re-uploading files that already exist with the same size.

### Force Overwrite
Use the `Force` flag in `ActionParams` to overwrite existing files regardless of size comparison.

### Share Expiration
Set expiration times for share links using the `Expires` field in `ShareParams`.

### Password Protection
Add password protection to share links using the `Password` field in `ShareParams`.

## Dependencies

- `github.com/duke-git/lancet/v2`: Utility functions for file operations
- `github.com/eventials/go-tus`: TUS protocol implementation for file uploads
- `github.com/imroc/req/v3`: HTTP client for API requests

## License

This project is licensed under the MIT License. 