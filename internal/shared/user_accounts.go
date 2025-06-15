package shared

// enum:oauth,credentials
type ProviderTypes string

const (
	ProviderTypeOAuth       ProviderTypes = "oauth"
	ProviderTypeCredentials ProviderTypes = "credentials"
)

func (p ProviderTypes) String() string {
	return string(p)
}

// enum:google,apple,facebook,github,credentials
type Providers string

const (
	ProvidersGoogle      Providers = "google"
	ProvidersApple       Providers = "apple"
	ProvidersFacebook    Providers = "facebook"
	ProvidersGithub      Providers = "github"
	ProvidersCredentials Providers = "credentials"
)

func (p Providers) String() string {
	return string(p)
}
