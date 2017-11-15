package aero

import "github.com/aerogo/http/client"

// Linter interface defines Begin and End methods that linters can implement.
type Linter interface {
	Begin(route string, uri string)
	End(route string, uri string, response client.Response)
}
