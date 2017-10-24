package aero

import "github.com/aerogo/http/client"

// Linter ...
type Linter interface {
	Begin(route string, uri string)
	End(route string, uri string, response client.Response)
}
