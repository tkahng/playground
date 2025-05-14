package oauth

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/tkahng/authgo/internal/conf"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type ProviderConfig interface {
	Pkce() bool
	Active() bool
	BuildAuthURL(state string, opts ...oauth2.AuthCodeOption) string
	Client(ctx context.Context, token *oauth2.Token) *http.Client
	FetchAuthUser(ctx context.Context, token *oauth2.Token) (*AuthUser, error)
	FetchRawUserInfo(ctx context.Context, token *oauth2.Token) ([]byte, error)
	FetchToken(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	FetchTokenOptions(verifier string) []oauth2.AuthCodeOption
}

// func

func OAuth2ConfigFromEnv(cfg conf.EnvConfig) {
	if cfg.GithubClientId != "" && cfg.GithubClientSecret != "" {
		Providers[NameGithub] = wrapFactory(func() ProviderConfig {
			return &GithubConfig{
				OAuth2ProviderConfig{
					ClientID:     cfg.GithubClientId,
					ClientSecret: cfg.GithubClientSecret,
					Name:         "GitHub",
					Enabled:      true,
					PKCE:         true, // technically is not supported yet but it is safe as the PKCE params are just ignored
					Scopes:       []string{"read:user", "user:email"},
					AuthURL:      github.Endpoint.AuthURL,
					TokenURL:     github.Endpoint.TokenURL,
					UserInfoURL:  "https://api.github.com/user",
					RedirectURL:  cfg.AppConfig.AppUrl + cfg.OAuth2Config.AuthCallback,
				},
			}
		})

	}
	if cfg.GoogleClientId != "" && cfg.GoogleClientSecret != "" {
		Providers[NameGoogle] = wrapFactory(func() ProviderConfig {
			return &GoogleConfig{
				OAuth2ProviderConfig: OAuth2ProviderConfig{
					Name:         "Google",
					ClientID:     cfg.GoogleClientId,
					ClientSecret: cfg.GoogleClientSecret,
					Enabled:      true,
					PKCE:         true,
					Scopes: []string{
						"https://www.googleapis.com/auth/userinfo.profile",
						"https://www.googleapis.com/auth/userinfo.email",
					},
					AuthURL:     "https://accounts.google.com/o/oauth2/v2/auth",
					TokenURL:    "https://oauth2.googleapis.com/token",
					UserInfoURL: "https://www.googleapis.com/oauth2/v3/userinfo",
					RedirectURL: cfg.AppConfig.AppUrl + cfg.OAuth2Config.AuthCallback,
				}}
		})
	}
}

const NameGithub = "github"
const NameGoogle = "google"

type OAuth2ProviderConfig struct {
	ClientID     string
	ClientSecret string
	Enabled      bool
	AuthURL      string
	TokenURL     string
	PKCE         bool
	UserInfoURL  string
	Name         string
	Scopes       []string
	RedirectURL  string
}

func (c OAuth2ProviderConfig) Pkce() bool {
	return c.PKCE
}

func (p *OAuth2ProviderConfig) oauth2Config() *oauth2.Config {
	return &oauth2.Config{
		RedirectURL:  p.RedirectURL,
		ClientID:     p.ClientID,
		ClientSecret: p.ClientSecret,
		Scopes:       p.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  p.AuthURL,
			TokenURL: p.TokenURL,
		},
	}
}

func (p *OAuth2ProviderConfig) FetchTokenOptions(verifier string) []oauth2.AuthCodeOption {
	var opts []oauth2.AuthCodeOption = []oauth2.AuthCodeOption{
		oauth2.AccessTypeOffline,
	}

	if p.Pkce() {
		opts = append(opts, oauth2.SetAuthURLParam("code_verifier", verifier))
	}
	return opts
}

// BuildAuthURL implements Provider.BuildAuthURL() interface method.
func (p *OAuth2ProviderConfig) BuildAuthURL(state string, opts ...oauth2.AuthCodeOption) string {
	return p.oauth2Config().AuthCodeURL(state, opts...)
}

// FetchToken implements Provider.FetchToken() interface method.
func (p *OAuth2ProviderConfig) FetchToken(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return p.oauth2Config().Exchange(ctx, code, opts...)
}

// Client implements Provider.Client() interface method.
func (p *OAuth2ProviderConfig) Client(ctx context.Context, token *oauth2.Token) *http.Client {
	return p.oauth2Config().Client(ctx, token)
}

// FetchRawUserInfo implements Provider.FetchRawUserInfo() interface method.
func (p *OAuth2ProviderConfig) FetchRawUserInfo(ctx context.Context, token *oauth2.Token) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}

	return p.sendRawUserInfoRequest(ctx, req, token)
}

// sendRawUserInfoRequest sends the specified user info request and return its raw response body.
func (p *OAuth2ProviderConfig) sendRawUserInfoRequest(ctx context.Context, req *http.Request, token *oauth2.Token) ([]byte, error) {
	client := p.Client(ctx, token)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	result, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// http.Client.Get doesn't treat non 2xx responses as error
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf(
			"failed to fetch OAuth2 user profile via %s (%d):\n%s",
			p.UserInfoURL,
			res.StatusCode,
			string(result),
		)
	}

	return result, nil
}
