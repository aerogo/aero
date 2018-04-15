package aero

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/aerogo/csp"

	"crypto/sha256"

	"encoding/base64"

	"github.com/aerogo/http/client"
	performance "github.com/aerogo/linter-performance"
	"github.com/aerogo/session"
	memstore "github.com/aerogo/session-store-memory"
	"github.com/fatih/color"
	"github.com/julienschmidt/httprouter"
)

// Application represents a single web service.
type Application struct {
	Config                *Configuration
	Sessions              session.Manager
	Security              ApplicationSecurity
	Servers               [2]*http.Server
	Linters               []Linter
	Router                *httprouter.Router
	ContentSecurityPolicy *csp.ContentSecurityPolicy

	root           string
	routeTests     map[string][]string
	start          time.Time
	rewrite        func(*RewriteContext)
	middleware     []Middleware
	pushConditions []func(*Context) bool
	onShutdown     []func()
	onPush         []func(*Context)

	routes struct {
		GET  []string
		POST []string
	}

	css            string
	cssHash        string
	cssReplacement string
}

// New creates a new application.
func New() *Application {
	app := new(Application)
	app.start = time.Now()
	app.routeTests = make(map[string][]string)
	app.Router = httprouter.New()

	// Default linters
	app.Linters = []Linter{
		performance.New(),
	}

	// Default CSP
	app.ContentSecurityPolicy = csp.New()
	app.ContentSecurityPolicy.SetMap(map[string]string{
		"default-src":  "'none'",
		"img-src":      "https:",
		"media-src":    "https:",
		"script-src":   "'self'",
		"style-src":    "'self'",
		"font-src":     "https:",
		"manifest-src": "'self'",
		"connect-src":  "https: wss:",
		"worker-src":   "'self'",
		"frame-src":    "https:",
	})

	// Configuration
	app.Config = new(Configuration)
	app.Config.Reset()
	app.Load()

	// Default session store: Memory
	app.Sessions.Store = memstore.New()

	// Default style
	// app.SetStyle("")

	// Set mime type for WebP because Go standard library doesn't include it
	mime.AddExtensionType(".webp", "image/webp")

	return app
}

// Get registers your function to be called when a certain GET path has been requested.
func (app *Application) Get(path string, handle Handle) {
	app.routes.GET = append(app.routes.GET, path)
	app.Router.GET(path, app.createRouteHandler(path, handle))
}

// Post registers your function to be called when a certain POST path has been requested.
func (app *Application) Post(path string, handle Handle) {
	app.routes.POST = append(app.routes.POST, path)
	app.Router.POST(path, app.createRouteHandler(path, handle))
}

// createRouteHandler creates a handler function for httprouter.
func (app *Application) createRouteHandler(path string, handle Handle) httprouter.Handle {
	return func(response http.ResponseWriter, request *http.Request, params httprouter.Params) {
		// Create context.
		ctx := Context{
			App:        app,
			StatusCode: http.StatusOK,
			request:    request,
			response:   response,
			params:     params,
		}

		// The last part of the call chain will send the actual response.
		lastPartOfCallChain := func() {
			data := handle(&ctx)
			ctx.respond(data)
		}

		// Declare the type of generateNext so that we can define it recursively in the next part.
		var generateNext func(index int) func()

		// Create a function that returns a bound function next()
		// which can be used as the 2nd parameter in the call chain.
		generateNext = func(index int) func() {
			if index == len(app.middleware) {
				return lastPartOfCallChain
			}

			return func() {
				app.middleware[index](&ctx, generateNext(index+1))
			}
		}

		generateNext(0)()
	}
}

// Run starts your application.
func (app *Application) Run() {
	app.TestManifest()
	app.TestRoutes()
	app.Listen()
	app.Wait()
	app.Shutdown()
}

// Use adds middleware to your middleware chain.
func (app *Application) Use(middlewares ...Middleware) {
	app.middleware = append(app.middleware, middlewares...)
}

// Load loads the application configuration from config.json.
func (app *Application) Load() {
	config, err := LoadConfig("config.json")

	if err != nil {
		// Ignore missing config file, we can perfectly run without one
		return
	}

	app.Config = config
}

// Listen starts the server.
func (app *Application) Listen() {
	if app.Security.Key != "" && app.Security.Certificate != "" {
		go func() {
			app.serveHTTPS(":" + strconv.Itoa(app.Config.Ports.HTTPS))
		}()

		fmt.Println("Server running on:", color.GreenString("https://localhost:"+strconv.Itoa(app.Config.Ports.HTTPS)))
	} else {
		fmt.Println("Server running on:", color.GreenString("http://localhost:"+strconv.Itoa(app.Config.Ports.HTTP)))
	}

	go func() {
		app.serveHTTP(":" + strconv.Itoa(app.Config.Ports.HTTP))
	}()
}

// Wait will make the process wait until it is killed.
func (app *Application) Wait() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-stop
}

// Shutdown will gracefully shut down the server.
func (app *Application) Shutdown() {
	for _, server := range app.Servers {
		if server == nil {
			continue
		}

		err := server.Shutdown(context.Background())

		if err != nil {
			fmt.Println(err)
		}
	}

	for _, callback := range app.onShutdown {
		callback()
	}
}

// OnShutdown registers a callback to be executed on server shutdown.
func (app *Application) OnShutdown(callback func()) {
	app.onShutdown = append(app.onShutdown, callback)
}

// OnPush registers a callback to be executed when an HTTP/2 push happens.
func (app *Application) OnPush(callback func(*Context)) {
	app.onPush = append(app.onPush, callback)
}

// AddPushCondition registers a callback to be executed when an HTTP/2 push happens.
func (app *Application) AddPushCondition(test func(*Context) bool) {
	app.pushConditions = append(app.pushConditions, test)
}

// Rewrite sets the URL rewrite function.
func (app *Application) Rewrite(rewrite func(*RewriteContext)) {
	app.rewrite = rewrite
}

// SetStyle applies the given CSS code to the application.
func (app *Application) SetStyle(css string) {
	app.css = css

	// Generate a hash
	hash := sha256.Sum256([]byte(css))
	app.cssHash = base64.StdEncoding.EncodeToString(hash[:])

	// Content security policy
	app.ContentSecurityPolicy.Set("style-src", "'sha256-"+app.cssHash+"'")

	// This will be used in the final response later on to inject the CSS code
	app.cssReplacement = "<style>" + app.css + "</style></head><body"
}

// Finalize post-processes the HTML to add styles to the output.
func (app *Application) Finalize(html string) string {
	if app.cssReplacement == "" {
		return html
	}

	return strings.Replace(html, "</head><body", app.cssReplacement, 1)
}

// StartTime returns the time the application started.
func (app *Application) StartTime() time.Time {
	return app.start
}

// Handler returns the request handler used by the application.
func (app *Application) Handler() http.Handler {
	router := app.Router
	rewrite := app.rewrite

	if rewrite != nil {
		return &rewriteHandler{
			rewrite: rewrite,
			router:  router,
		}
	}

	return router
}

// createServer creates an http server instance.
func (app *Application) createServer(address string) *http.Server {
	return &http.Server{
		Addr:              address,
		Handler:           app.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      180 * time.Second,
		IdleTimeout:       120 * time.Second,
		TLSConfig:         createTLSConfig(),
	}
}

// serveHTTP serves requests from the given listener.
func (app *Application) serveHTTP(address string) {
	server := app.createServer(address)
	app.Servers[0] = server

	// This will block the calling goroutine until the server shuts down.
	serveError := server.ListenAndServe()

	if serveError != nil && strings.Index(serveError.Error(), "closed") == -1 {
		panic(serveError)
	}
}

// serveHTTPS serves requests from the given listener.
func (app *Application) serveHTTPS(address string) {
	server := app.createServer(address)
	app.Servers[1] = server

	// This will block the calling goroutine until the server shuts down.
	serveError := server.ListenAndServeTLS(app.Security.Certificate, app.Security.Key)

	if serveError != nil && strings.Index(serveError.Error(), "closed") == -1 {
		panic(serveError)
	}
}

// Test tests the given URI paths when the application starts.
func (app *Application) Test(route string, paths []string) {
	app.routeTests[route] = paths
}

// TestManifest tests your application's manifest.
func (app *Application) TestManifest() {
	manifest := app.Config.Manifest

	// Warn about short name length (Google Lighthouse)
	// https://developer.chrome.com/apps/manifest/name#short_name
	if len(manifest.ShortName) >= 12 {
		color.Yellow("The short name of your application should have less than 12 characters")
	}
}

// TestRoutes tests your application's routes.
func (app *Application) TestRoutes() {
	fmt.Println(strings.Repeat("-", 80))

	go func() {
		sort.Strings(app.routes.GET)

		for _, route := range app.routes.GET {
			// Skip ajax routes
			if strings.HasPrefix(route, "/_") {
				continue
			}

			// Check if the user defined test routes for the given route
			testRoutes, exists := app.routeTests[route]

			if exists {
				for _, testRoute := range testRoutes {
					app.TestRoute(route, testRoute)
				}

				continue
			}

			// Skip routes with parameters and display a warning to indicate it needs a test route
			if strings.Contains(route, ":") {
				color.Yellow(route)
				continue
			}

			// Test the static route without parameters
			app.TestRoute(route, route)
		}

		// json, _ := Post("https://html5.validator.nu/?out=json").Header("Content-Type", "text/html; charset=utf-8").Header("Content-Encoding", "gzip").Body(body).Send()
		// fmt.Println(json)
	}()
}

// TestRoute tests the given route.
func (app *Application) TestRoute(route string, uri string) {
	for _, linter := range app.Linters {
		linter.Begin(route, uri)
	}

	response, _ := client.Get("http://localhost:" + strconv.Itoa(app.Config.Ports.HTTP) + uri).End()

	for _, linter := range app.Linters {
		linter.End(route, uri, response)
	}
}
