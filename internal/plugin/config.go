package plugin

type Config struct {
	RedirectsAppURL string   `json:"redirectsAppURL,omitempty"`
	V2              V2Config `json:"v2,omitempty"`
}

type V2Config struct {
	Enabled      bool   `json:"enabled,omitempty"`
	ClientName   string `json:"clientName,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty"`
	ServerURL    string `json:"serverURL,omitempty"`
	JwtSecret    string `json:"jwtSecret,omitempty"`
}
