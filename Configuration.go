package aero

// Configuration ...
type Configuration struct {
	Domain    string   `json:"domain"`
	Title     string   `json:"title"`
	Fonts     []string `json:"fonts"`
	Icons     []string `json:"icons"`
	Static    []string `json:"static"`
	Styles    []string `json:"styles"`
	GZip      bool     `json:"gzip"`
	GZipCache bool     `json:"gzipCache"`
	Manifest  struct {
		GCMSenderID string `json:"gcm_sender_id"`
	} `json:"manifest"`
	Ports struct {
		HTTP  int `json:"http"`
		HTTPS int `json:"https"`
	} `json:"ports"`
}

// Reset resets all fields to the default configuration.
func (config *Configuration) Reset() {
	config.GZip = true
	config.GZipCache = true
	config.Ports.HTTP = 4000
	config.Ports.HTTPS = 4001
}
