package aero

import (
	"compress/gzip"
	stdContext "context"
	"fmt"
	"io"
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
	"github.com/akyoto/color"
)

// Application represents a single web service.
type Application struct {
	Config                *Configuration
	Sessions              session.Manager
	Security              ApplicationSecurity
	Linters               []Linter
	ContentSecurityPolicy *csp.ContentSecurityPolicy

	router         Router
	routeTests     map[string][]string
	start          time.Time
	rewrite        []func(RewriteContext)
	middleware     []Middleware
	pushConditions []func(Context) bool
	onStart        []func()
	onShutdown     []func()
	onPush         []func(Context)
	onError        []func(Context, error)
	stop           chan os.Signal
	pushOptions    http.PushOptions
	contextPool    sync.Pool
	gzipWriterPool sync.Pool
	serversMutex   sync.Mutex
	servers        [2]*http.Server

	routes struct {
		GET []string
	}
}

// New creates a new application.
func New() *Application {
	app := &Application{
		start:                 time.Now(),
		stop:                  make(chan os.Signal, 1),
		routeTests:            make(map[string][]string),
		Config:                &Configuration{},
		ContentSecurityPolicy: csp.New(),

		// Default linters
		Linters: []Linter{
			performance.New(),
		},
	}

	// Default CSP
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

	// Context pool
	app.contextPool.New = func() interface{} {
		return &context{
			app: app,
		}
	}

	// Push options describes the headers that are sent
	// to our server to retrieve the push response.
	app.pushOptions = http.PushOptions{
		Method: "GET",
		Header: http.Header{
			acceptEncodingHeader: []string{"gzip"},
		},
	}

	// Configuration
	app.Config.Reset()
	app.Load()

	// Default session store: Memory
	app.Sessions.Store = memstore.New()

	// MIME types
	initMIMETypes()

	// Receive signals
	signal.Notify(app.stop, os.Interrupt, syscall.SIGTERM)

	return app
}

// Get registers your function to be called when the given GET path has been requested.
func (app *Application) Get(path string, handler Handler) {
	app.routes.GET = append(app.routes.GET, path)
	app.router.Add(http.MethodGet, path, handler)
}

// Post registers your function to be called when the given POST path has been requested.
func (app *Application) Post(path string, handler Handler) {
	app.router.Add(http.MethodPost, path, handler)
}

// Delete registers your function to be called when the given DELETE path has been requested.
func (app *Application) Delete(path string, handler Handler) {
	app.router.Add(http.MethodDelete, path, handler)
}

// Put registers your function to be called when the given PUT path has been requested.
func (app *Application) Put(path string, handler Handler) {
	app.router.Add(http.MethodPut, path, handler)
}

// Any registers your function to be called with any http method.
func (app *Application) Any(path string, handler Handler) {
	app.Get(path, handler)
	app.Post(path, handler)
	app.Delete(path, handler)
	app.Put(path, handler)
}

// Router returns the router used by the application.
func (app *Application) Router() *Router {
	return &app.router
}

// Run starts your application.
func (app *Application) Run() {
	app.BindMiddleware()
	app.ListenAndServe()

	for _, callback := range app.onStart {
		callback()
	}

	app.TestRoutes()
	app.wait()
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
		go app.serveHTTPS(listener)
		fmt.Println("Server running on:", color.GreenString("https://localhost:"+strconv.Itoa(app.Config.Ports.HTTPS)))
	}

	listener := app.listen(":" + strconv.Itoa(app.Config.Ports.HTTP))
	go app.serveHTTP(listener)
	fmt.Println("Server running on:", color.GreenString("http://localhost:"+strconv.Itoa(app.Config.Ports.HTTP)))
}

// wait will make the process wait until it is killed.
func (app *Application) wait() {
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

// OnStart registers a callback to be executed on server start.
func (app *Application) OnStart(callback func()) {
	app.onStart = append(app.onStart, callback)
}

// OnEnd registers a callback to be executed on server shutdown.
func (app *Application) OnEnd(callback func()) {
	app.onShutdown = append(app.onShutdown, callback)
}

// OnPush registers a callback to be executed when an HTTP/2 push happens.
func (app *Application) OnPush(callback func(Context)) {
	app.onPush = append(app.onPush, callback)
}

// OnError registers a callback to be executed on server errors.
func (app *Application) OnError(callback func(Context, error)) {
	app.onError = append(app.onError, callback)
}

// AddPushCondition registers a callback that
// needs to return true before an HTTP/2 push happens.
func (app *Application) AddPushCondition(test func(Context) bool) {
	app.pushConditions = append(app.pushConditions, test)
}

// Rewrite adds a URL path rewrite function.
func (app *Application) Rewrite(rewrite func(RewriteContext)) {
	app.rewrite = append(app.rewrite, rewrite)
}

// StartTime returns the time the application started.
func (app *Application) StartTime() time.Time {
	return app.start
}

// NewContext returns a new context from the pool.
func (app *Application) NewContext(req *http.Request, res http.ResponseWriter) *context {
	ctx := app.contextPool.Get().(*context)
	ctx.status = http.StatusOK
	ctx.request = request{inner: req}
	ctx.response = response{inner: res}
	ctx.session = nil
	ctx.paramCount = 0
	return ctx
}

// ServeHTTP responds to the given request.
func (app *Application) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	ctx := app.NewContext(request, response)

	for _, rewrite := range app.rewrite {
		rewrite(ctx)
	}

	app.router.Lookup(request.Method, request.URL.Path, ctx)

	if ctx.handler == nil {
		response.WriteHeader(http.StatusNotFound)
		ctx.Close()
		return
	}

	err := ctx.handler(ctx)

	if err != nil {
		for _, callback := range app.onError {
			callback(ctx, err)
		}
	}

	ctx.Close()
}

// Test tests the given URI paths when the application starts.
func (app *Application) Test(route string, paths ...string) {
	app.routeTests[route] = paths
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

// acquireGZipWriter will return a clean gzip writer from the pool.
func (app *Application) acquireGZipWriter(response io.Writer) *gzip.Writer {
	var writer *gzip.Writer
	obj := app.gzipWriterPool.Get()

	if obj == nil {
		writer, _ = gzip.NewWriterLevel(response, gzip.BestCompression)
		return writer
	}

	writer = obj.(*gzip.Writer)
	writer.Reset(response)
	return writer
}

// BindMiddleware applies the middleware to every router node.
// This is called by `Run` automatically and should never be called
// outside of tests.
func (app *Application) BindMiddleware() {
	app.router.Each(func(node *tree) {
		if node.data != nil {
			node.data = app.chain(node.data)
		}
	})
}

// chain chains all the middleware and returns a new handler.
func (app *Application) chain(handler Handler) Handler {
	for i := len(app.middleware) - 1; i >= 0; i-- {
		handler = app.middleware[i](handler)
	}

	return handler
}

// createServer creates an http server instance.
func (app *Application) createServer() *http.Server {
	return &http.Server{
		Handler:           app,
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
	// The returned error is never nil and in case of a normal shutdown
	// it will be `http.ErrServerClosed`.
	err := server.Serve(listener)

	if err != http.ErrServerClosed {
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
	// The returned error is never nil and in case of a normal shutdown
	// it will be `http.ErrServerClosed`.
	err := server.ServeTLS(listener, app.Security.Certificate, app.Security.Key)

	if err != http.ErrServerClosed {
		panic(err)
	}
}

// shutdown will gracefully shut down the server.
func shutdown(server *http.Server) {
	if server == nil {
		return
	}

	// Add a timeout to the server shutdown
	ctx, cancel := stdContext.WithTimeout(stdContext.Background(), 250*time.Millisecond)
	defer cancel()

	// Shut down server
	err := server.Shutdown(ctx)

	if err != nil {
		fmt.Println(err)
	}
}

// initMIMETypes adds a few additional types to the MIME package.
func initMIMETypes() {
	mimeTypes := []struct {
		extension string
		typ       string
	}{
		{
			extension: ".webp",
			typ:       "image/webp",
		},
		{
			extension: ".apng",
			typ:       "image/apng",
		},
	}

	for _, m := range mimeTypes {
		err := mime.AddExtensionType(m.extension, m.typ)

		if err != nil {
			color.Red("Failed adding '%s' MIME extension", m.typ)
		}
	}
}
