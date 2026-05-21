package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// BackblazeService handles file uploads to Backblaze B2
type BackblazeService struct {
	KeyID          string
	ApplicationKey string
	BucketID       string
	BucketName     string
	AuthToken      string
	APIURL         string
	DownloadURL    string
	UploadURL      string
	UploadToken    string
}

type b2AuthResponse struct {
	AccountID          string `json:"accountId"`
	AuthorizationToken string `json:"authorizationToken"`
	APIURL             string `json:"apiUrl"`
	DownloadURL        string `json:"downloadUrl"`
}

type b2UploadURLResponse struct {
	BucketID           string `json:"bucketId"`
	UploadURL          string `json:"uploadUrl"`
	AuthorizationToken string `json:"authorizationToken"`
}

type b2UploadResponse struct {
	FileID   string `json:"fileId"`
	FileName string `json:"fileName"`
}

// NewBackblazeService creates and authorizes a new Backblaze service instance.
// Returns nil (not an error) when credentials are absent so the app still starts
// without B2 configured.
func NewBackblazeService() (*BackblazeService, error) {
	keyID := os.Getenv("BACKBLAZE_KEY_ID")
	appKey := os.Getenv("BACKBLAZE_APPLICATION_KEY")
	bucketID := os.Getenv("BACKBLAZE_BUCKET_ID")
	bucketName := os.Getenv("BACKBLAZE_BUCKET_NAME")

	if keyID == "" || appKey == "" || bucketID == "" || bucketName == "" {
		return nil, fmt.Errorf("identifiants Backblaze manquants dans les variables d'environnement")
	}

	svc := &BackblazeService{
		KeyID:          keyID,
		ApplicationKey: appKey,
		BucketID:       bucketID,
		BucketName:     bucketName,
	}

	if err := svc.authorize(); err != nil {
		return nil, err
	}

	return svc, nil
}

// authorize authenticates with Backblaze B2 API
func (b *BackblazeService) authorize() error {
	req, err := http.NewRequest("GET", "https://api.backblazeb2.com/b2api/v2/b2_authorize_account", nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(b.KeyID, b.ApplicationKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("autorisation B2 échouée: %s", string(body))
	}

	var authResp b2AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return err
	}

	b.AuthToken = authResp.AuthorizationToken
	b.APIURL = authResp.APIURL
	b.DownloadURL = authResp.DownloadURL
	return nil
}

// getUploadURL fetches a one-time upload URL from B2
func (b *BackblazeService) getUploadURL() error {
	url := b.APIURL + "/b2api/v2/b2_get_upload_url"
	payload := map[string]string{"bucketId": b.BucketID}
	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", b.AuthToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("impossible d'obtenir l'URL d'upload B2: %s", string(body))
	}

	var uploadResp b2UploadURLResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return err
	}

	b.UploadURL = uploadResp.UploadURL
	b.UploadToken = uploadResp.AuthorizationToken
	return nil
}

// UploadFile uploads any file to Backblaze B2 under the given folder prefix
func (b *BackblazeService) UploadFile(file *multipart.FileHeader, folder string) (string, error) {
	if b.UploadURL == "" || b.UploadToken == "" {
		if err := b.getUploadURL(); err != nil {
			return "", err
		}
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	fileBytes, err := io.ReadAll(src)
	if err != nil {
		return "", err
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	filename := fmt.Sprintf("%s/%s%s", folder, uuid.New().String(), ext)

	req, err := http.NewRequest("POST", b.UploadURL, bytes.NewBuffer(fileBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", b.UploadToken)
	req.Header.Set("X-Bz-File-Name", filename)
	req.Header.Set("Content-Type", file.Header.Get("Content-Type"))
	req.Header.Set("X-Bz-Content-Sha1", "do_not_verify")
	req.ContentLength = int64(len(fileBytes))

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		// Token may have expired — re-authorize and retry once
		if authErr := b.authorize(); authErr != nil {
			return "", authErr
		}
		if urlErr := b.getUploadURL(); urlErr != nil {
			return "", urlErr
		}
		req2, _ := http.NewRequest("POST", b.UploadURL, bytes.NewBuffer(fileBytes))
		req2.Header.Set("Authorization", b.UploadToken)
		req2.Header.Set("X-Bz-File-Name", filename)
		req2.Header.Set("Content-Type", file.Header.Get("Content-Type"))
		req2.Header.Set("X-Bz-Content-Sha1", "do_not_verify")
		req2.ContentLength = int64(len(fileBytes))
		resp, err = client.Do(req2)
		if err != nil {
			return "", err
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload B2 échoué: %s", string(body))
	}

	fileURL := fmt.Sprintf("%s/file/%s/%s", b.DownloadURL, b.BucketName, filename)
	return fileURL, nil
}

// UploadImage uploads an image to the "images" folder in B2
func (b *BackblazeService) UploadImage(file *multipart.FileHeader) (string, error) {
	return b.UploadFile(file, "images")
}
