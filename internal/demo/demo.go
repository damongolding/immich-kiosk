package demo

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/charmbracelet/log"
)

type LoginResponse struct {
	AccessToken          string `json:"accessToken"`
	UserID               string `json:"userId"`
	UserEmail            string `json:"userEmail"`
	Name                 string `json:"name"`
	IsAdmin              bool   `json:"isAdmin"`
	ProfileImagePath     string `json:"profileImagePath"`
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

var demoTokenMutex sync.RWMutex
var DemoToken string

func ValidateToken(ctx context.Context, token string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://demo.immich.app/api/auth/validateToken", nil)
	if err != nil {
		log.Error("ValidateToken: new request", "err", err)
		return false
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("ValidateToken: sendingrequest", "err", err)
		return false
	}
	defer resp.Body.Close()

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

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://demo.immich.app/api/auth/login", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error("DemoLogin: creating request", "err", err)
		return "", err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Create HTTP client and send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("DemoLogin: sending request", "err", err)
		return "", err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("DemoLogin: reading request", "err", err)
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
