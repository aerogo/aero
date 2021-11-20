package aero

import (
	"compress/gzip"
	stdContext "context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/aerogo/csp"
	"github.com/aerogo/session"
	memstore "github.com/aerogo/session-store-memory"
	"github.com/akyoto/color"
)

// Application represents a single web service.
type Application struct {
	Config                *Configuration
	Sessions              session.Manager
	Security              ApplicationSecurity
	ContentSecurityPolicy *csp.ContentSecurityPolicy

	router         Router
	rewrite        []func(RewriteContext)
	middleware     []Middleware
	pushConditions []func(Context) bool
	contextPool    sync.Pool
	gzipWriterPool sync.Pool
	pushOptions    http.PushOptions
	serversMutex   sync.Mutex
	servers        [2]*http.Server
	stop           chan os.Signal

	onStart    []func()
	onShutdown []func()
	onPush     []func(Context)
	onError    []func(Context, error)
}

// New creates a new application.
func New() *Application {
	app := &Application{
		Config:                &Configuration{},
		ContentSecurityPolicy: csp.New(),
		stop:                  make(chan os.Signal, 1),
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

	// MIME types
	_ = mime.AddExtensionType(".apng", "image/apng")

	// Default SameSite value is "Lax"
	app.Sessions.SameSite = http.SameSiteLaxMode

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

	// Default session store: Memory
	app.Sessions.Store = memstore.New()

	// Configuration
	app.Config.Reset()
	app.Load()

	return app
}

// Get registers your function to be called when the given GET path has been requested.
func (app *Application) Get(path string, handler Handler) {
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
	signal.Notify(app.stop, os.Interrupt, syscall.SIGTERM)
	app.BindMiddleware()
	app.ListenAndServe()

	for _, callback := range app.onStart {
		callback()
	}

	<-app.stop
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

// Shutdown will gracefully shut down all servers.
func (app *Application) Shutdown() {
	app.serversMutex.Lock()
	defer app.serversMutex.Unlock()

	for _, server := range app.servers {
		shutdown(server, app.Config.Timeouts.Shutdown)
	}

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

// newContext returns a new context from the pool.
func (app *Application) newContext(req *http.Request, res http.ResponseWriter) *context {
	ctx := app.contextPool.Get().(*context)
	ctx.status = http.StatusOK
	ctx.request.inner = req
	ctx.response.inner = res
	ctx.session = nil
	ctx.paramCount = 0
	ctx.modifierCount = 0
	return ctx
}

// ServeHTTP responds to the given request.
func (app *Application) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	ctx := app.newContext(request, response)

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
	app.router.bind(func(handler Handler) Handler {
		return handler.Bind(app.middleware...)
	})
}

// createServer creates an http server instance.
func (app *Application) createServer() *http.Server {
	return &http.Server{
		Handler:           app,
		ReadHeaderTimeout: app.Config.Timeouts.ReadHeader,
		WriteTimeout:      app.Config.Timeouts.Write,
		IdleTimeout:       app.Config.Timeouts.Idle,
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

	if !errors.Is(err, http.ErrServerClosed) {
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

	if !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}

// shutdown will gracefully shut down the server.
func shutdown(server *http.Server, timeout time.Duration) {
	if server == nil {
		return
	}

	// Add a timeout to the server shutdown
	ctx, cancel := stdContext.WithTimeout(stdContext.Background(), timeout)
	defer cancel()

	// Shut down server
	err := server.Shutdown(ctx)

	if err != nil {
		fmt.Println(err)
	}
}
