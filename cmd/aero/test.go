package main

const mainTestCode = `package main

import (
	"net/http/httptest"
	"testing"

	"github.com/aerogo/aero"
)

func TestStaticRoutes(t *testing.T) {
	app := configure(aero.New())
	request := httptest.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("Invalid status %d", response.Code)
	}
}
`
