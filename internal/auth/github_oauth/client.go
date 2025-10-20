package github_oauth

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

type Client struct {
	oauthClientId string
	oauthSecret   string
	logger        *slog.Logger
}

type oauthTokenRequest struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Code         string `json:"code"`
}

func (c *Client) GetToken(code string) (string, error) {
	data := oauthTokenRequest{
		ClientId:     c.oauthClientId,
		ClientSecret: c.oauthSecret,
		Code:         code,
	}

	jsonData, err := json.Marshal(data)

	if err != nil {
		return "", err
	}

	c.logger.Debug("github oauth token request", "data", string(jsonData))

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	c.logger.Debug("github oauth token response", "body", string(body), "status-code", resp.StatusCode)

	// Optional: decode into a struct
	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", err
	}

	return tokenResponse.AccessToken, nil
}

func NewClient(oauthClientId string, oauthSecret string, logger *slog.Logger) *Client {
	return &Client{
		oauthClientId: oauthClientId,
		oauthSecret:   oauthSecret,
		logger:        logger,
	}
}
