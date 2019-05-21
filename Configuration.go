package aero

import (
	"os"

	jsoniter "github.com/json-iterator/go"
)

// Configuration represents the data in your config.json file.
type Configuration struct {
	Domain   string            `json:"domain"`
	Title    string            `json:"title"`
	Push     []string          `json:"push"`
	GZip     bool              `json:"gzip"`
	Manifest Manifest          `json:"manifest"`
	Ports    PortConfiguration `json:"ports"`
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
}

// ManifestIcon represents a single icon in the web manifest.
type ManifestIcon struct {
	Source string `json:"src"`
	Sizes  string `json:"sizes"`
	Type   string `json:"type"`
}

// PortConfiguration lets you configure the ports that Aero will listen on.
type PortConfiguration struct {
	HTTP  int `json:"http"`
	HTTPS int `json:"https"`
}

// Reset resets all fields to the default configuration.
func (config *Configuration) Reset() {
	config.GZip = true
	config.Ports.HTTP = 4000
	config.Ports.HTTPS = 4001
	config.Manifest.StartURL = "/"
	config.Manifest.Display = "standalone"
	config.Manifest.Lang = "en"
	config.Manifest.ShortName = "Untitled"
	config.Title = "Untitled site"
	config.Push = []string{}
}

// LoadConfig loads the application configuration from the file system.
func LoadConfig(path string) (*Configuration, error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer file.Close()
	config := &Configuration{}
	config.Reset()

	decoder := jsoniter.NewDecoder(file)
	err = decoder.Decode(config)

	if err != nil {
		return nil, err
	}

	if config.Manifest.Name == "" {
		config.Manifest.Name = config.Title
	}

	if config.Manifest.ShortName == "" {
		config.Manifest.ShortName = config.Title
	}

	return config, nil
}
