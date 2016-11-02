package aero

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/buaazp/fasthttprouter"
	"github.com/fatih/color"
	"github.com/valyala/fasthttp"
)

// Application represents a single web service.
type Application struct {
	Config   Configuration
	Security struct {
		Key         []byte
		Certificate []byte
	}

	root   string
	router *fasthttprouter.Router
}

// New creates a new application.
func New() *Application {
	app := new(Application)
	app.root = ""
	app.router = fasthttprouter.New()
	app.Config.Reset()

	return app
}

// Get registers your function to be called when a certain path has been requested.
func (app *Application) Get(path string, handle Handle) {
	app.router.GET(path, func(fasthttpContext *fasthttp.RequestCtx, params fasthttprouter.Params) {
		ctx := Context{
			App:        app,
			Params:     params,
			requestCtx: fasthttpContext,
			start:      time.Now(),
		}

		handle(&ctx)
	})
}

// Run calls app.Load() and app.Listen().
func (app *Application) Run() {
	app.Load()
	app.Listen()
}

// Load loads the application data from the file system.
func (app *Application) Load() {
	// TODO: ...
}

// Listen starts the server.
func (app *Application) Listen() {
	fmt.Println("Server running on:", color.GreenString("http://localhost:"+strconv.Itoa(app.Config.Ports.HTTP)))

	listener := app.listen()
	app.serve(listener)
}

// listen listens on the specified host and port.
func (app *Application) listen() net.Listener {
	address := ":" + strconv.Itoa(app.Config.Ports.HTTP)

	listener, bindError := net.Listen("tcp", address)

	if bindError != nil {
		panic(bindError)
	}

	return listener
}

// serve serves requests from the given listener.
func (app *Application) serve(listener net.Listener) {
	server := &fasthttp.Server{
		Handler: app.router.Handler,
	}

	if app.Security.Key != nil && app.Security.Certificate != nil {
		serveError := server.ServeTLSEmbed(listener, app.Security.Certificate, app.Security.Key)

		if serveError != nil {
			panic(serveError)
		}
	}

	serveError := server.Serve(listener)

	if serveError != nil {
		panic(serveError)
	}
}
