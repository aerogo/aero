package aero

import "crypto/tls"

// ApplicationSecurity stores the certificate data.
type ApplicationSecurity struct {
	Key         string
	Certificate string
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
		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{
			tls.CurveP521,
			tls.CurveP384,
			tls.CurveP256,
		},
		CipherSuites: []uint16{
			// ECDSA is about 3 times faster than RSA on the server side.
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,

			// RSA is slower on the server side but still widely used.
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}
}
