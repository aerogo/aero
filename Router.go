package aero

import (
	"fmt"
	"net/http"
	"os"
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
func (router *Router) Add(method string, path string, handler Handler) {
	tree := router.selectTree(method)

	if tree == nil {
		panic(fmt.Errorf("Unknown HTTP method: '%s'", method))
	}

	tree.add(path, handler)
}

// Find returns the handler for the given route.
// This is only useful for testing purposes.
// Use Lookup instead.
func (router *Router) Find(method string, path string) Handler {
	c := context{}
	router.Lookup(method, path, &c)
	return c.handler
}

// Lookup finds the handler and parameters for the given route
// and assigns them to the given context.
func (router *Router) Lookup(method string, path string, ctx *context) {
	if method[0] == 'G' {
		router.get.find(path, ctx)
		return
	}

	tree := router.selectTree(method)
	tree.find(path, ctx)
}

// bind traverses all trees and calls the given function on every node.
func (router *Router) bind(transform func(Handler) Handler) {
	router.get.bind(transform)
	router.post.bind(transform)
	router.delete.bind(transform)
	router.put.bind(transform)
	router.patch.bind(transform)
	router.head.bind(transform)
	router.connect.bind(transform)
	router.trace.bind(transform)
	router.options.bind(transform)
}

// Print shows a pretty print of the dynamic routes.
func (router *Router) Print(method string) {
	tree := router.selectTree(method)
	tree.root.PrettyPrint(os.Stdout)
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
