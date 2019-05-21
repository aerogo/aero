package aero

import (
	"mime"
	"os"
)

func init() {
	// Set mime type for WebP because the Go standard library doesn't include it
	err := mime.AddExtensionType(".webp", "image/webp")

	if err != nil {
		os.Stderr.WriteString("Failed adding image/webp MIME extension")
	}

	// Set mime type for APNG because the one in Go differs
	err = mime.AddExtensionType(".apng", "image/apng")

	if err != nil {
		os.Stderr.WriteString("Failed adding image/apng MIME extension")
	}
}
