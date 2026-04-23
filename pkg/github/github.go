// Package github provides utilities for interacting with the GitHub API
// and constructing MCP tools for GitHub operations.
package github

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// ClientConfig holds configuration for creating a GitHub client.
type ClientConfig struct {
	// Token is the GitHub personal access token or app token.
	Token string
	// BaseURL is the optional GitHub Enterprise base URL.
	// If empty, the public GitHub API is used.
	BaseURL string
	// UploadURL is the optional GitHub Enterprise upload URL.
	UploadURL string
}

// NewClient creates a new GitHub API client using the provided configuration.
// If a token is provided, the client will authenticate with it.
// If BaseURL is set, the client will connect to a GitHub Enterprise instance.
func NewClient(ctx context.Context, cfg ClientConfig) (*github.Client, error) {
	var httpClient *http.Client

	if cfg.Token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: cfg.Token},
		)
		httpClient = oauth2.NewClient(ctx, ts)
	}

	if cfg.BaseURL != "" {
		uploadURL := cfg.UploadURL
		if uploadURL == "" {
			uploadURL = cfg.BaseURL
		}
		client, err := github.NewEnterpriseClient(cfg.BaseURL, uploadURL, httpClient)
		if err != nil {
			return nil, fmt.Errorf("failed to create GitHub Enterprise client: %w", err)
		}
		return client, nil
	}

	if httpClient != nil {
		return github.NewClient(httpClient), nil
	}

	return github.NewClient(nil), nil
}

// GetAuthenticatedUser retrieves the currently authenticated GitHub user.
// Returns an error if the client is not authenticated or the request fails.
func GetAuthenticatedUser(ctx context.Context, client *github.Client) (*github.User, error) {
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticated user: %w", err)
	}
	return user, nil
}

// IsTokenValid checks whether the provided GitHub token is valid by attempting
// to retrieve the authenticated user. Returns true if the token is valid.
func IsTokenValid(ctx context.Context, token string) bool {
	client, err := NewClient(ctx, ClientConfig{Token: token})
	if err != nil {
		return false
	}
	_, err = GetAuthenticatedUser(ctx, client)
	return err == nil
}
