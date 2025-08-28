package demo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/log"
)

type LoginResponse struct {
	AccessToken          string `json:"accessToken"`
	UserID               string `json:"userId"`
	UserEmail            string `json:"userEmail"`
	Name                 string `json:"name"`
	ProfileImagePath     string `json:"profileImagePath"`
	IsAdmin              bool   `json:"isAdmin"`
	ShouldChangePassword bool   `json:"shouldChangePassword"`
}

type APIKey struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Permissions []string  `json:"permissions"`
}

type ValidateResponse struct {
	AuthStatus bool `json:"authStatus"`
}

const demoImmichURL = "https://demo.immich.app"

var demoTokenMutex sync.RWMutex
var DemoToken string

var (
	httpClient = &http.Client{
		Timeout: 30 * time.Second,
	}
)

func ValidateToken(ctx context.Context, token string) bool {

	demoURL := demoImmichURL
	if os.Getenv("KIOSK_IMMICH_URL") != "" {
		demoURL = os.Getenv("KIOSK_IMMICH_URL")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, demoURL+"/api/auth/validateToken", nil)
	if err != nil {
		log.Error("ValidateToken: new request", "err", err)
		return false
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Error("ValidateToken: sending request", "err", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		log.Warn("ValidateToken: token is invalid", "status", resp.StatusCode)
		return false
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Error("ValidateToken: unexpected status code", "status", resp.StatusCode)
		return false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("ValidateToken: reading response", "err", err)
		return false
	}

	// Parse JSON response into struct
	var demoValidateResponse ValidateResponse
	err = json.Unmarshal(body, &demoValidateResponse)
	if err != nil {
		log.Error("ValidateToken: parsing JSON response", "err", err)
		return false
	}

	return demoValidateResponse.AuthStatus
}

func Login(ctx context.Context, refresh bool) (string, error) {
	demoTokenMutex.RLock()

	if DemoToken != "" && !refresh {
		defer demoTokenMutex.RUnlock()
		return DemoToken, nil
	}

	demoTokenMutex.RUnlock()

	payload := map[string]string{
		"email":    "demo@immich.app",
		"password": "demo",
	}

	// Convert payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Error("DemoLogin: Error marshalling JSON", "err", err)
		return "", err
	}

	demoURL := demoImmichURL
	if os.Getenv("KIOSK_IMMICH_URL") != "" {
		demoURL = os.Getenv("KIOSK_IMMICH_URL")
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, demoURL+"/api/auth/login", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error("DemoLogin: creating request", "err", err)
		return "", err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Create HTTP client and send request
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Error("DemoLogin: sending request", "err", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Error("DemoLogin: unexpected status code", "status", resp.StatusCode)
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("DemoLogin: reading response", "err", err)
		return "", err
	}

	// Parse JSON response into struct
	var loginResp LoginResponse
	err = json.Unmarshal(body, &loginResp)
	if err != nil {
		log.Error("DemoLogin: parsing JSON response", "err", err)
		return "", err
	}

	demoTokenMutex.Lock()
	DemoToken = loginResp.AccessToken
	demoTokenMutex.Unlock()

	log.Debug("Retrieved demo token", "token", DemoToken)

	return DemoToken, nil
}
