package aero

import "errors"

var (
	ErrAddressNotValid            = errors.New("Address is not valid")
	ErrEmptyBody                  = errors.New("Empty body")
	ErrExpectedJSONObject         = errors.New("Invalid format: Expected JSON object")
	ErrRequestInterruptedByClient = errors.New("Request interrupted by the client")
)
