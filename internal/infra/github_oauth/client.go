package github_oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/google/go-github/v69/github"
	"github.com/panoptescloud/api/internal/domain"
	"github.com/panoptescloud/api/internal/domain/users"
	"golang.org/x/oauth2"
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

	if resp.StatusCode != 200 {
		return "", errors.New("unexpected response from github oauth token request")
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
		AccessToken      string `json:"access_token"`
		TokenType        string `json:"token_type"`
		Scope            string `json:"scope"`
		ErrorCode        string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}

	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", err
	}

	if tokenResponse.ErrorCode != "" {
		if tokenResponse.ErrorCode == "bad_verification_code" {
			return "", domain.ErrUnauthorised{
				Message: tokenResponse.ErrorDescription,
			}
		}

		return "", fmt.Errorf("failed to get github token (%s): %s", tokenResponse.ErrorCode, tokenResponse.ErrorDescription)
	}

	return tokenResponse.AccessToken, nil
}

func (c *Client) GetProfile(accessToken string) (users.GithubProfile, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	ghClient := github.NewClient(tc)

	// empty string means "authenticated user"
	user, _, err := ghClient.Users.Get(ctx, "")

	if err != nil {
		return users.GithubProfile{}, err
	}

	return users.GithubProfile{
		NodeID: user.GetNodeID(),
		Name:   user.GetName(),
		Email:  user.GetEmail(),
	}, nil
}

func NewClient(oauthClientId string, oauthSecret string, logger *slog.Logger) *Client {
	return &Client{
		oauthClientId: oauthClientId,
		oauthSecret:   oauthSecret,
		logger:        logger,
	}
}
