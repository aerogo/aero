package aero

import (
	"fmt"
	"net/http"
)

// Router is a high-performance router.
type Router struct {
	get    tree
	post   tree
	delete tree
}

// tree represents a radix tree.
type tree struct {
	prefix   string
	handle   Handle
	children []*tree
}

// prettyPrint
func (node *tree) prettyPrint() {
	fmt.Println(node.prefix, node.handle, len(node.children))

	for _, child := range node.children {
		child.prettyPrint()
	}
}

// add adds a new element to the tree.
func (node *tree) add(path string, handle Handle) {
	// Search tree for equal parts until we can no longer proceed
	i := 0
	offset := 0

	for {
		// The path already exists.
		// node: /blog|
		// path: /blog|
		if i == len(path) {
			node.handle = handle
			return
		}

		// The node we just checked is completely included in our path.
		// node: /|
		// path: /|blog
		if i == len(node.prefix) {
			// Try to search children
			for _, child := range node.children {
				if child.prefix[0] == path[i] {
					offset = i
					node = child
					goto next
				}
			}

			// No fitting children found, does this node even contain a prefix yet?
			// If no prefix is set, this is the starting node.
			if node.prefix == "" {
				node.prefix = path
				node.handle = handle
				return
			}

			// Otherwise, add a new child with the remaining string.
			node.children = append(node.children, &tree{
				prefix: path[i:],
				handle: handle,
			})
			return
		}

		// We got a conflict.
		// node: /b|ag
		// path: /b|riefcase
		if path[i] != node.prefix[i-offset] {
			// Create split node with the remaining string
			splitNode := &tree{
				prefix:   node.prefix[i-offset:],
				handle:   node.handle,
				children: node.children,
			}

			// Create added node with the remaining string in the path
			addedNode := &tree{
				prefix: path[i:],
				handle: handle,
			}

			// Cut the existing node
			node.prefix = node.prefix[:i-offset]
			node.handle = nil
			node.children = []*tree{
				splitNode,
				addedNode,
			}
			return
		}

	next:
		i++
	}
}

// Add registers a new handler for the given method and path.
func (router *Router) Add(method string, path string, handle Handle) {
	tree := router.selectTree(method)
	tree.add(path, handle)
}

// Find returns the handle for the given route.
func (router *Router) Find(method string, path string) Handle {
	tree := router.selectTree(method)
	tree.prettyPrint()

	if tree.prefix == path {
		return tree.handle
	}

	return nil
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
