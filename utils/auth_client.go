package utils

import (
	"encoding/json"
	"fmt"
	"net/http"

	"gitlab.com/nodiviti/user-service/config"
)

type AuthClient struct {
	baseURL    string
	httpClient *http.Client
}

type ValidateTokenResponse struct {
	Valid bool `json:"valid"`
	User  struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	} `json:"user"`
}

func NewAuthClient(cfg *config.Config) *AuthClient {
	return &AuthClient{
		baseURL: cfg.AuthService.URL,
		httpClient: &http.Client{
			Timeout: cfg.AuthService.Timeout,
		},
	}
}

// ValidateToken validates JWT token with auth service
func (c *AuthClient) ValidateToken(token string) (*ValidateTokenResponse, error) {
	url := fmt.Sprintf("%s/api/v1/auth/validate", c.baseURL)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid token: status %d", resp.StatusCode)
	}

	var validateResp ValidateTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&validateResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &validateResp, nil
}

// GetUserInfo gets basic user info from auth service
func (c *AuthClient) GetUserInfo(userID int) (*ValidateTokenResponse, error) {
	// This would be a separate endpoint in auth service to get user by ID
	// For now, we'll implement token validation only
	return nil, fmt.Errorf("not implemented yet")
}
