package aero

// Configuration ...
type Configuration struct {
	GZip      bool
	GZipCache bool
	Ports     struct {
		HTTP  int
		HTTPS int
	}
}

// Reset resets all fields to the default configuration.
func (config *Configuration) Reset() {
	config.GZip = true
	config.GZipCache = true
	config.Ports.HTTP = 4000
	config.Ports.HTTPS = 4001
}
