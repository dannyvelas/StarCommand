package resolve

type rootConfig struct {
	BitwardenAPIURL      string `yaml:"bitwarden_api_url" json:"bitwarden_api_url"`
	BitwardenIdentityURL string `yaml:"bitwarden_identity_url" json:"bitwarden_identity_url"`
}

var defaultRootConfig = rootConfig{
	BitwardenAPIURL:      "https://api.bitwarden.com",
	BitwardenIdentityURL: "https://identity.bitwarden.com",
}
