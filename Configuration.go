package aero

import (
	"os"

	jsoniter "github.com/json-iterator/go"
)

// Configuration represents the data in your config.json file.
type Configuration struct {
	Domain string            `json:"domain"`
	Title  string            `json:"title"`
	Push   []string          `json:"push"`
	GZip   bool              `json:"gzip"`
	Ports  PortConfiguration `json:"ports"`
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

	return config, nil
}
