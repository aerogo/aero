package aero

import (
	"net/http"
	"strings"
)

// Router is a high-performance router.
type Router struct {
	get    tree
	post   tree
	delete tree
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
		return tree.handle
	}

	return tree.find(path)
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
	default:
		return nil
	}
}
