package aero

// Router is a high-performance router.
type Router struct {
	get node
}

type node struct {
	prefix   string
	handle   Handle
	children []*node
}

// Add registers a new handler for the given method and path.
func (router *Router) Add(method string, path string, handle Handle) {
	// path = strings.TrimPrefix(path, "/")
	// router.get = append(router.get.children, &node{
	// 	prefix: path,
	// 	handle: handle,
	// })
	router.get.prefix = path
	router.get.handle = handle
}

// Find ...
func (router *Router) Find(method string, path string) Handle {
	return router.get.handle
}

// Exec responds to the given request.
func (router *Router) Exec(method string, path string, ctx *Context) {
	tree := router.get
	tree.handle(ctx)
}
