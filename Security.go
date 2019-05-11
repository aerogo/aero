package aero

import (
	"crypto/tls"

	"github.com/aerogo/http/ciphers"
)

// ApplicationSecurity stores the certificate data.
type ApplicationSecurity struct {
	Certificate string
	Key         string
}

// Load expects the path of the certificate and the key.
func (security *ApplicationSecurity) Load(certificate string, key string) {
	security.Certificate = certificate
	security.Key = key
}

// createTLSConfig creates a secure TLS configuration.
func createTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CipherSuites:             ciphers.List,
		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{
			tls.CurveP521,
			tls.CurveP384,
			tls.CurveP256,
		},
	}
}
