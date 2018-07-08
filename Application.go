package aero

import (
	"context"
	"fmt"
	"mime"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/aerogo/csp"

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
	Linters               []Linter
	Router                *httprouter.Router
	ContentSecurityPolicy *csp.ContentSecurityPolicy

	servers        [2]*http.Server
	serversMutex   sync.Mutex
	routeTests     map[string][]string
	start          time.Time
	rewrite        func(*RewriteContext)
	middleware     []Middleware
	pushConditions []func(*Context) bool
	onStart        []func()
	onShutdown     []func()
	onPush         []func(*Context)
	stop           chan os.Signal

	routes struct {
		GET  []string
		POST []string
	}
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
	app.ContentSecurityPolicy.SetMap(csp.Map{
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
		"base-uri":     "'self'",
		"form-action":  "'self'",
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

	// Receive signals
	app.stop = make(chan os.Signal, 1)
	signal.Notify(app.stop, os.Interrupt, syscall.SIGTERM)

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
	app.ListenAndServe()

	for _, callback := range app.onStart {
		callback()
	}

	app.TestManifest()
	app.TestRoutes()
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

// ListenAndServe starts the server.
// It guarantees that a TCP listener is listening on the ports defined in the config
// when the function returns.
func (app *Application) ListenAndServe() {
	if app.Security.Key != "" && app.Security.Certificate != "" {
		listener := app.listen(":" + strconv.Itoa(app.Config.Ports.HTTPS))

		go func() {
			app.serveHTTPS(listener)
		}()

		fmt.Println("Server running on:", color.GreenString("https://localhost:"+strconv.Itoa(app.Config.Ports.HTTPS)))
	} else {
		fmt.Println("Server running on:", color.GreenString("http://localhost:"+strconv.Itoa(app.Config.Ports.HTTP)))
	}

	listener := app.listen(":" + strconv.Itoa(app.Config.Ports.HTTP))

	go func() {
		app.serveHTTP(listener)
	}()
}

// Wait will make the process wait until it is killed.
func (app *Application) Wait() {
	<-app.stop
}

// Shutdown will gracefully shut down all servers.
func (app *Application) Shutdown() {
	app.serversMutex.Lock()
	defer app.serversMutex.Unlock()

	shutdown(app.servers[0])
	shutdown(app.servers[1])

	for _, callback := range app.onShutdown {
		callback()
	}
}

// shutdown will gracefully shut down the server.
func shutdown(server *http.Server) {
	if server == nil {
		return
	}

	// Add a timeout to the server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	// Shut down server
	err := server.Shutdown(ctx)

	if err != nil {
		fmt.Println(err)
	}
}

// OnStart registers a callback to be executed on server start.
func (app *Application) OnStart(callback func()) {
	app.onStart = append(app.onStart, callback)
}

// OnEnd registers a callback to be executed on server shutdown.
func (app *Application) OnEnd(callback func()) {
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
func (app *Application) createServer() *http.Server {
	return &http.Server{
		Handler:           app.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      180 * time.Second,
		IdleTimeout:       120 * time.Second,
		TLSConfig:         createTLSConfig(),
	}
}

// listen returns a Listener for the given address.
func (app *Application) listen(address string) Listener {
	listener, err := net.Listen("tcp", address)

	if err != nil {
		panic(err)
	}

	return Listener{listener.(*net.TCPListener)}
}

// serveHTTP serves requests from the given listener.
func (app *Application) serveHTTP(listener Listener) {
	server := app.createServer()

	app.serversMutex.Lock()
	app.servers[0] = server
	app.serversMutex.Unlock()

	// This will block the calling goroutine until the server shuts down.
	err := server.Serve(listener)

	if err != nil && !strings.Contains(err.Error(), "closed") {
		panic(err)
	}
}

// serveHTTPS serves requests from the given listener.
func (app *Application) serveHTTPS(listener Listener) {
	server := app.createServer()

	app.serversMutex.Lock()
	app.servers[1] = server
	app.serversMutex.Unlock()

	// This will block the calling goroutine until the server shuts down.
	err := server.ServeTLS(listener, app.Security.Certificate, app.Security.Key)

	if err != nil && !strings.Contains(err.Error(), "closed") {
		panic(err)
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
