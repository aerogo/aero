package aero

import (
	"encoding/json"
	"os"
)

// Configuration represents the data in your config.json file.
type Configuration struct {
	Push  []string          `json:"push"`
	GZip  bool              `json:"gzip"`
	Ports PortConfiguration `json:"ports"`
}

// PortConfiguration lets you configure the ports that Aero will listen on.
type PortConfiguration struct {
	HTTP  int `json:"http"`
	HTTPS int `json:"https"`
}

// Reset resets all fields to the default configuration.
func (config *Configuration) Reset() {
	config.Push = []string{}
	config.GZip = true
	config.Ports.HTTP = 4000
	config.Ports.HTTPS = 4001
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

	decoder := json.NewDecoder(file)
	err = decoder.Decode(config)

	if err != nil {
		return nil, err
	}

	return config, nil
}
