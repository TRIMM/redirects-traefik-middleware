package plugin

type Config struct {
	RedirectsAppURL string `json:"redirectsAppURL,omitempty"`
	V2              bool   `json:"v2,omitempty"`
}
