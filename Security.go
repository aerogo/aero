package aero

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
