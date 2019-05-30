package aero

import (
	"net/http"
	"strings"
)

// Router is a high-performance router.
type Router struct {
	get     tree
	post    tree
	delete  tree
	put     tree
	patch   tree
	head    tree
	connect tree
	trace   tree
	options tree
}

// Add registers a new handler for the given method and path.
func (router *Router) Add(method string, path string, handle Handle) {
	tree := router.selectTree(method)
	tree.add(path, handle)
}

// Find returns the handle for the given route.
func (router *Router) Find(method string, path string) Handle {
	tree := router.selectTree(method)

	// Fast path for the root node
	if tree.prefix == path {
		handle, _ := tree.data.(Handle)
		return handle
	}

	handle, _ := tree.find(path).(Handle)
	return handle
}

// String returns a pretty print of the GET routes.
func (router *Router) String() string {
	buffer := strings.Builder{}
	router.get.prettyPrint(&buffer)
	return buffer.String()
}

// selectTree returns the tree by the given HTTP method.
func (router *Router) selectTree(method string) *tree {
	switch method {
	case http.MethodGet:
		return &router.get
	case http.MethodPost:
		return &router.post
	case http.MethodDelete:
		return &router.delete
	case http.MethodPut:
		return &router.put
	case http.MethodPatch:
		return &router.patch
	case http.MethodHead:
		return &router.head
	case http.MethodConnect:
		return &router.connect
	case http.MethodTrace:
		return &router.trace
	case http.MethodOptions:
		return &router.options
	default:
		return nil
	}
}
