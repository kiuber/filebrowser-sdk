package filebrowser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocalPathForDownload(t *testing.T) {
	tests := []struct {
		name     string
		fileURL  string
		expected string
	}{
		{
			name:     "Simple URL with path",
			fileURL:  "https://example.com/files/document.pdf",
			expected: filepath.Join(os.TempDir(), "files", "document.pdf"),
		},
		{
			name:     "URL with root path",
			fileURL:  "https://example.com/document.pdf",
			expected: filepath.Join(os.TempDir(), "document.pdf"),
		},
		{
			name:     "URL with empty path",
			fileURL:  "https://example.com/",
			expected: filepath.Join(os.TempDir(), "downloaded_file"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LocalPathForDownload(tt.fileURL)
			if result != tt.expected {
				t.Errorf("LocalPathForDownload() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEnsureFolderForFile(t *testing.T) {
	// Test with valid path
	testPath := filepath.Join(os.TempDir(), "test", "folder", "file.txt")
	err := EnsureFolderForFile(testPath)
	if err != nil {
		t.Errorf("EnsureFolderForFile() error = %v", err)
	}

	// Verify directory was created
	dir := filepath.Dir(testPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("Directory was not created: %s", dir)
	}

	// Test with empty path
	err = EnsureFolderForFile("")
	if err == nil {
		t.Error("EnsureFolderForFile() should return error for empty path")
	}

	// Cleanup
	os.RemoveAll(filepath.Join(os.TempDir(), "test"))
}

func TestFileExistsWithSameSize(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "test-file")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write some content to get a known size
	content := []byte("test content")
	_, err = tempFile.Write(content)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// Test with correct size
	exists := fileExistsWithSameSize(tempFile.Name(), int64(len(content)))
	if !exists {
		t.Error("fileExistsWithSameSize() should return true for existing file with correct size")
	}

	// Test with wrong size
	exists = fileExistsWithSameSize(tempFile.Name(), int64(len(content)+1))
	if exists {
		t.Error("fileExistsWithSameSize() should return false for existing file with wrong size")
	}

	// Test with non-existent file
	exists = fileExistsWithSameSize("non-existent-file.txt", 100)
	if exists {
		t.Error("fileExistsWithSameSize() should return false for non-existent file")
	}
}

func TestFilebrowserAuthValidate(t *testing.T) {
	tests := []struct {
		name    string
		auth    FilebrowserAuth
		wantErr bool
	}{
		{
			name: "Valid auth",
			auth: FilebrowserAuth{
				URL:      "https://example.com",
				Username: "user",
				Password: "pass",
			},
			wantErr: false,
		},
		{
			name: "Empty URL",
			auth: FilebrowserAuth{
				URL:      "",
				Username: "user",
				Password: "pass",
			},
			wantErr: true,
		},
		{
			name: "Empty username",
			auth: FilebrowserAuth{
				URL:      "https://example.com",
				Username: "",
				Password: "pass",
			},
			wantErr: true,
		},
		{
			name: "Empty password",
			auth: FilebrowserAuth{
				URL:      "https://example.com",
				Username: "user",
				Password: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.auth.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("FilebrowserAuth.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
} 