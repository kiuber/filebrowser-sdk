package filebrowser

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/eventials/go-tus"
	"github.com/imroc/req/v3"
)

// Client represents a Filebrowser client
type Client struct {
	URL string
	ReqLogin
	Token string
}

// ReqLogin contains login request parameters
type ReqLogin struct {
	Username string
	Password string
}

// ReqShare contains share request parameters
type ReqShare struct {
	Expires  string
	Password string
	Unit     string
}

// RespLogin contains login response data
type RespLogin struct {
	Token string
}

// RespResource contains resource information
type RespResource struct {
	NotExist  bool   `json:"not_exist"`
	Path      string `json:"path"`
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	Extension string `json:"extension"`
	Modified  string `json:"modified"`
	Mode      int64  `json:"mode"`
	IsDir     string `json:"IsDir"`
	IsSymlink string `json:"isSymlink"`
	Type      string `json:"type"`
}

// RespShare contains share response data
type RespShare struct {
	Hash string `json:"hash"`
	Path string `json:"path"`
}

// Validate checks if the client configuration is valid
func (c *Client) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("URL cannot be empty")
	}
	if c.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if c.Password == "" {
		return fmt.Errorf("password cannot be empty")
	}
	return nil
}

// Login authenticates with the Filebrowser server and retrieves a token
func (c *Client) Login() error {
	if err := c.Validate(); err != nil {
		return fmt.Errorf("invalid client configuration: %w", err)
	}

	client := req.C().DevMode()
	resp, err := client.R().
		SetBody(ReqLogin{Username: c.Username, Password: c.Password}).
		Post(fmt.Sprintf("%s/api/login", c.URL))
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status code: %d", resp.StatusCode)
	}

	c.Token = resp.String()
	if c.Token == "" {
		return fmt.Errorf("received empty token from server")
	}

	log.Printf("Successfully authenticated with Filebrowser")
	return nil
}

// ensureAuthenticated ensures the client is authenticated, logging in if necessary
func (c *Client) ensureAuthenticated() error {
	if c.Token == "" {
		return c.Login()
	}
	return nil
}

// Upload uploads a local file to the specified remote path using TUS protocol
func (c *Client) Upload(localPath string, remotePath string) error {
	if localPath == "" {
		return fmt.Errorf("local path cannot be empty")
	}
	if remotePath == "" {
		return fmt.Errorf("remote path cannot be empty")
	}

	// Check if local file exists
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return fmt.Errorf("local file does not exist: %s", localPath)
	}

	if err := c.ensureAuthenticated(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Configure TUS client
	config := tus.DefaultConfig()
	config.Header.Set("X-Auth", c.Token)
	
	tusClient, err := tus.NewClient(
		fmt.Sprintf("%s/api/tus/%s", c.URL, remotePath),
		config,
	)
	if err != nil {
		return fmt.Errorf("failed to create TUS client: %w", err)
	}

	// Open local file
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	// Create upload from file
	upload, err := tus.NewUploadFromFile(file)
	if err != nil {
		return fmt.Errorf("failed to create upload from file: %w", err)
	}

	// Create uploader
	uploader, err := tusClient.CreateUpload(upload)
	if err != nil {
		return fmt.Errorf("failed to create upload: %w", err)
	}

	// Perform upload
	if err := uploader.Upload(); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	log.Printf("Successfully uploaded file to remote path: %s", remotePath)
	return nil
}

// Share creates a share link for the specified remote path
func (c *Client) Share(remotePath string, expires int64, password string, unit string) (string, error) {
	if remotePath == "" {
		return "", fmt.Errorf("remote path cannot be empty")
	}

	if err := c.ensureAuthenticated(); err != nil {
		return "", fmt.Errorf("authentication failed: %w", err)
	}

	// Prepare share request body
	body := ReqShare{}
	if expires > 0 {
		body = ReqShare{
			Expires:  fmt.Sprintf("%d", expires),
			Password: password,
			Unit:     unit,
		}
	}

	// Make share request
	var result RespShare
	client := req.C()
	resp, err := client.R().
		SetHeader("X-Auth", c.Token).
		SetBody(body).
		SetSuccessResult(&result).
		Post(fmt.Sprintf("%s/api/share/%s", c.URL, remotePath))
	if err != nil {
		return "", fmt.Errorf("share request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("share request failed with status code: %d", resp.StatusCode)
	}

	if result.Hash == "" {
		return "", fmt.Errorf("received empty hash from server")
	}

	log.Printf("Successfully created share for path: %s", remotePath)
	return result.Hash, nil
}

// GetResource retrieves information about a resource at the specified path
func (c *Client) GetResource(remotePath string) (*RespResource, error) {
	if remotePath == "" {
		return nil, fmt.Errorf("remote path cannot be empty")
	}

	if err := c.ensureAuthenticated(); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Make resource request
	var result RespResource
	client := req.C()
	url := fmt.Sprintf("%s/api/resources/%s", c.URL, remotePath)
	resp, err := client.R().
		SetHeader("X-Auth", c.Token).
		SetSuccessResult(&result).
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("resource request failed: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return &RespResource{NotExist: true}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("resource request failed with status code: %d", resp.StatusCode)
	}

	return &result, nil
}

// DeleteResource deletes a resource at the specified path
func (c *Client) DeleteResource(remotePath string) error {
	if remotePath == "" {
		return fmt.Errorf("remote path cannot be empty")
	}

	if err := c.ensureAuthenticated(); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Make delete request
	client := req.C()
	url := fmt.Sprintf("%s/api/resources/%s", c.URL, remotePath)
	resp, err := client.R().
		SetHeader("X-Auth", c.Token).
		Delete(url)
	if err != nil {
		return fmt.Errorf("delete request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("delete request failed with status code: %d", resp.StatusCode)
	}

	log.Printf("Successfully deleted resource: %s", remotePath)
	return nil
}
