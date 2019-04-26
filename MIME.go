package aero

import "mime"

func init() {
	// Set mime type for WebP because the Go standard library doesn't include it
	mime.AddExtensionType(".webp", "image/webp")
}
