package immich

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
)

type DemoLoginResponse struct {
	AccessToken          string `json:"accessToken"`
	UserID               string `json:"userId"`
	UserEmail            string `json:"userEmail"`
	Name                 string `json:"name"`
	IsAdmin              bool   `json:"isAdmin"`
	ProfileImagePath     string `json:"profileImagePath"`
	ShouldChangePassword bool   `json:"shouldChangePassword"`
}

type DemoAPIKey struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Permissions []string  `json:"permissions"`
}

type DemoValidateResponse struct {
	AuthStatus bool `json:"authStatus"`
}

var demoToken string

func validateDemoToken(token string) bool {
	req, err := http.NewRequest("POST", "https://demo.immich.app/api/auth/validateToken", nil)
	if err != nil {
		return false
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return false
	}

	// Parse JSON response into struct
	var demoValidateResponse DemoValidateResponse
	err = json.Unmarshal(body, &demoValidateResponse)
	if err != nil {
		fmt.Printf("Error parsing JSON response: %v\n", err)
		return false
	}

	return demoValidateResponse.AuthStatus
}

func demoLogin(refresh bool) (string, error) {

	if demoToken != "" && !refresh {
		return demoToken, nil
	}

	payload := map[string]string{
		"email":    "demo@immich.app",
		"password": "demo",
	}

	// Convert payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshalling JSON: %v\n", err)
		return "", err
	}

	// Create the request
	req, err := http.NewRequest("POST", "https://demo.immich.app/api/auth/login", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return "", err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Create HTTP client and send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return "", err
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return "", err
	}

	// Parse JSON response into struct
	var loginResp DemoLoginResponse
	err = json.Unmarshal(body, &loginResp)
	if err != nil {
		fmt.Printf("Error parsing JSON response: %v\n", err)
		return "", err
	}

	demoToken = loginResp.AccessToken

	log.Debug("Retrieved demo token", "token", demoToken)

	return demoToken, nil
}
