package aero

import (
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

// App ...
type App struct {
	root   string
	router *fasthttprouter.Router
}

// New ...
func New() *App {
	app := new(App)
	app.root = ""
	app.router = fasthttprouter.New()
	return app
}

// Get ...
func (app *App) Get(route string, handle fasthttprouter.Handle) {
	app.router.GET(route, handle)
}

// Run ...
func (app *App) Run() {
	fasthttp.ListenAndServe(":5000", app.router.Handler)
}
