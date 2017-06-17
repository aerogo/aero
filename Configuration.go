package aero

// Configuration ...
type Configuration struct {
	Domain    string            `json:"domain"`
	Title     string            `json:"title"`
	Fonts     []string          `json:"fonts"`
	Icons     []string          `json:"icons"`
	Static    []string          `json:"static"`
	Styles    []string          `json:"styles"`
	GZip      bool              `json:"gzip"`
	GZipCache bool              `json:"gzipCache"`
	Manifest  Manifest          `json:"manifest"`
	Ports     PortConfiguration `json:"ports"`
}

// Manifest represents a web manifest
type Manifest struct {
	Name            string         `json:"name"`
	ShortName       string         `json:"short_name"`
	Icons           []ManifestIcon `json:"icons,omitempty"`
	StartURL        string         `json:"start_url"`
	Display         string         `json:"display"`
	Lang            string         `json:"lang,omitempty"`
	ThemeColor      string         `json:"theme_color,omitempty"`
	BackgroundColor string         `json:"background_color,omitempty"`
	GCMSenderID     string         `json:"gcm_sender_id,omitempty"`
}

// ManifestIcon represents a single icon in the web manifest.
type ManifestIcon struct {
	Source string `json:"src"`
	Sizes  string `json:"sizes"`
}

// PortConfiguration ...
type PortConfiguration struct {
	HTTP  int `json:"http"`
	HTTPS int `json:"https"`
}

// Reset resets all fields to the default configuration.
func (config *Configuration) Reset() {
	config.GZip = true
	config.GZipCache = true
	config.Ports.HTTP = 4000
	config.Ports.HTTPS = 4001
	config.Title = "Untitled site"
}
