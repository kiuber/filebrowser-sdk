package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/kiuber/filebrowser-sdk"
)

func main() {
	// Example configuration - replace with your actual Filebrowser credentials
	auth := filebrowser.FilebrowserAuth{
		URL:      "https://your-filebrowser-instance.com",
		Username: "your-username",
		Password: "your-password",
	}

	// Example: Download and share a file
	actionParams := filebrowser.ActionParams{
		FileSize: 0, // Set to actual file size if known, 0 to skip size checking
		Force:    false,
		ShareParams: filebrowser.ShareParams{
			Expires:  24, // 24 hours
			Password: "",  // No password protection
			Unit:     "hours",
		},
	}

	// Function to generate remote path
	remotePathFn := func(filename string) string {
		return filepath.Join("uploads", "examples", filename)
	}

	// Example external file URL
	externalURL := "https://example.com/sample-file.pdf"

	fmt.Println("Starting file download and share process...")

	// Download, upload, and share the file
	result, err := filebrowser.SaveAndShare(
		auth,
		externalURL,
		remotePathFn,
		actionParams,
	)
	if err != nil {
		log.Fatalf("Failed to save and share file: %v", err)
	}

	fmt.Printf("✅ Success! File has been uploaded and shared.\n")
	fmt.Printf("📁 View URL: %s\n", result.ViewUrl)
	fmt.Printf("⬇️  Download URL: %s\n", result.DownloadUrl)

	// Example: Using the client directly for more control
	fmt.Println("\n--- Direct Client Usage Example ---")

	client := &filebrowser.Client{
		URL: auth.URL,
		ReqLogin: filebrowser.ReqLogin{
			Username: auth.Username,
			Password: auth.Password,
		},
	}

	// Login
	if err := client.Login(); err != nil {
		log.Fatalf("Failed to login: %v", err)
	}
	fmt.Println("✅ Successfully authenticated with Filebrowser")

	// Get resource info
	resource, err := client.GetResource("uploads/examples/sample-file.pdf")
	if err != nil {
		log.Printf("Warning: Failed to get resource info: %v", err)
	} else {
		fmt.Printf("📄 Resource info: %s (Size: %d bytes)\n", resource.Name, resource.Size)
	}

	// Create a share with password protection
	hash, err := client.Share("uploads/examples/sample-file.pdf", 7, "secret123", "days")
	if err != nil {
		log.Printf("Warning: Failed to create protected share: %v", err)
	} else {
		protectedViewURL := fmt.Sprintf("%s/share/%s", client.URL, hash)
		fmt.Printf("🔒 Protected share URL: %s\n", protectedViewURL)
	}
} 