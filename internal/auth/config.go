package auth

// Config holds authentication configuration
type Config struct {
	Strategy    string        `mapstructure:"strategy"`
	JWTSecret   string        `mapstructure:"jwt_secret"`
	TokenExpiry int           `mapstructure:"token_expiry"` // Hours
	GitHub      *GitHubConfig `mapstructure:"github"`
}

// GitHubConfig holds GitHub OAuth configuration
type GitHubConfig struct {
	Organization string   `mapstructure:"organization"`
	ClientID     string   `mapstructure:"client_id"`
	ClientSecret string   `mapstructure:"client_secret"`
	RedirectURL  string   `mapstructure:"redirect_url"`
	WriteTeams   []string `mapstructure:"write_teams"`
	Scopes       []string `mapstructure:"scopes"`
}
