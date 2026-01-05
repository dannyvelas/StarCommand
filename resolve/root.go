package resolve

type rootConfig struct {
	BitwardenAPIURL      string `json:"bitwarden_api_url"`
	BitwardenIdentityURL string `json:"bitwarden_identity_url"`
}

var defaultRootConfig = rootConfig{
	BitwardenAPIURL:      "https://api.bitwarden.com",
	BitwardenIdentityURL: "https://identity.bitwarden.com",
}
