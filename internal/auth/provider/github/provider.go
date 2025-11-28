package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v57/github"
	"github.com/mazurov/command-launcher-registry/internal/auth"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
)

// GitHubProvider implements AuthProvider for GitHub OAuth
type GitHubProvider struct {
	config      *auth.GitHubConfig
	oauthConfig *oauth2.Config
}

// NewGitHubProvider creates a new GitHub auth provider
func NewGitHubProvider(cfg *auth.GitHubConfig) *GitHubProvider {
	scopes := cfg.Scopes
	if len(scopes) == 0 {
		scopes = []string{"read:org", "user:email"}
	}

	return &GitHubProvider{
		config: cfg,
		oauthConfig: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       scopes,
			Endpoint:     githuboauth.Endpoint,
		},
	}
}

// Name returns provider identifier
func (p *GitHubProvider) Name() string {
	return "github"
}

// GetAuthURL returns OAuth authorization URL
func (p *GitHubProvider) GetAuthURL(state string) string {
	return p.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

// HandleCallback processes OAuth callback and returns authenticated user
func (p *GitHubProvider) HandleCallback(ctx context.Context, code string) (*auth.User, error) {
	// Exchange code for token
	token, err := p.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Create GitHub client
	client := github.NewClient(p.oauthConfig.Client(ctx, token))

	// Get user info
	githubUser, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get user's teams in the organization
	teams, err := p.getUserTeams(ctx, client, p.config.Organization)
	if err != nil {
		return nil, fmt.Errorf("failed to get teams: %w", err)
	}

	user := &auth.User{
		ID:        fmt.Sprintf("%d", githubUser.GetID()),
		Username:  githubUser.GetLogin(),
		Email:     githubUser.GetEmail(),
		Name:      githubUser.GetName(),
		AvatarURL: githubUser.GetAvatarURL(),
		Teams:     teams,
		Provider:  "github",
		CreatedAt: time.Now(),
	}

	return user, nil
}

// getUserTeams fetches user's teams from specified organization
func (p *GitHubProvider) getUserTeams(ctx context.Context, client *github.Client, org string) ([]string, error) {
	opts := &github.ListOptions{PerPage: 100}
	var allTeams []string

	for {
		teams, resp, err := client.Teams.ListUserTeams(ctx, opts)
		if err != nil {
			return nil, err
		}

		for _, team := range teams {
			// Only include teams from our organization
			if team.Organization != nil && team.Organization.GetLogin() == org {
				allTeams = append(allTeams, team.GetSlug())
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allTeams, nil
}

// GetUserTeamsWithClient fetches teams using provided client (for PAT exchange)
func (p *GitHubProvider) GetUserTeamsWithClient(ctx context.Context, client *github.Client, org string) ([]string, error) {
	return p.getUserTeams(ctx, client, org)
}

// ValidateToken validates JWT token and returns user info
func (p *GitHubProvider) ValidateToken(ctx context.Context, claims *auth.JWTClaims) (*auth.User, error) {
	// For JWT-based auth, we trust the claims
	// Optionally, could re-fetch teams from GitHub API for fresh data
	return &auth.User{
		ID:       claims.UserID,
		Username: claims.Username,
		Email:    claims.Email,
		Teams:    claims.Teams,
		Provider: claims.Provider,
	}, nil
}

// ValidatePAT validates GitHub Personal Access Token and returns user info
func (p *GitHubProvider) ValidatePAT(ctx context.Context, token string) (*auth.User, error) {
	// Create GitHub client with PAT
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Get user info
	githubUser, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get user with PAT: %w", err)
	}

	// Get user's teams in the organization
	teams, err := p.getUserTeams(ctx, client, p.config.Organization)
	if err != nil {
		return nil, fmt.Errorf("failed to get teams: %w", err)
	}

	user := &auth.User{
		ID:        fmt.Sprintf("%d", githubUser.GetID()),
		Username:  githubUser.GetLogin(),
		Email:     githubUser.GetEmail(),
		Name:      githubUser.GetName(),
		AvatarURL: githubUser.GetAvatarURL(),
		Teams:     teams,
		Provider:  "github",
		CreatedAt: time.Now(),
	}

	return user, nil
}

// Authenticate is not supported for GitHub provider (OAuth only)
func (p *GitHubProvider) Authenticate(ctx context.Context, username, password string) (*auth.User, error) {
	return nil, fmt.Errorf("direct authentication not supported for GitHub provider")
}
