package aero

import (
	"fmt"
	"io"

	"github.com/akyoto/color"
)

// tree represents a radix tree.
type tree struct {
	prefix   string
	handle   Handle
	children [256]*tree
}

// add adds a new element to the tree.
func (node *tree) add(path string, handle Handle) {
	// Search tree for equal parts until we can no longer proceed
	i := 0
	offset := 0

	for {
		if i == len(path) {
			// The path already exists.
			// node: /blog|
			// path: /blog|
			if i-offset == len(node.prefix) {
				node.handle = handle
				return
			}

			// The path ended but the node prefix is longer.
			// node: /blog|feed
			// path: /blog|
			node.split(i-offset, "", handle)
			return
		}

		// The node we just checked is entirely included in our path.
		// node: /|
		// path: /|blog
		if i-offset == len(node.prefix) {
			child := node.children[path[i]]

			if child != nil {
				offset = i
				node = child
				goto next
			}

			// No fitting children found, does this node even contain a prefix yet?
			// If no prefix is set, this is the starting node.
			if node.prefix == "" {
				node.prefix = path
				node.handle = handle
				return
			}

			// Otherwise, add a new child with the remaining string.
			node.children[path[i]] = &tree{
				prefix: path[i:],
				handle: handle,
			}

			return
		}

		// We got a conflict.
		// node: /b|ag
		// path: /b|riefcase
		if path[i] != node.prefix[i-offset] {
			node.split(i-offset, path[i:], handle)
			return
		}

	next:
		i++
	}
}

// split splits the node at the given index and inserts
// a new child node with the given path and handle.
// If path is empty, it will not create another child node
// and instead assign the handle directly to the node.
func (node *tree) split(index int, path string, handle Handle) {
	// Create split node with the remaining string
	splitNode := &tree{
		prefix:   node.prefix[index:],
		handle:   node.handle,
		children: node.children,
	}

	// Cut the existing node
	node.prefix = node.prefix[:index]

	// If the path is empty, it means we don't create a 2nd child node.
	// Just assign the handle for the existing node and store a single child node.
	if path == "" {
		node.handle = handle
		node.children[splitNode.prefix[0]] = splitNode
		return
	}

	// Create new node with the remaining string in the path
	newNode := &tree{
		prefix: path,
		handle: handle,
	}

	// The existing handle must be removed
	node.handle = nil

	// Assign new child nodes
	node.children = [256]*tree{}
	node.children[splitNode.prefix[0]] = splitNode
	node.children[newNode.prefix[0]] = splitNode
}

// find returns the handle for the given path, if available.
func (node *tree) find(path string) Handle {
	// Search tree for equal parts until we can no longer proceed
	i := 0
	offset := 0

	for {
		fmt.Println(i)

		// We reached the end.
		if i == len(path) {
			// node: /blog|
			// path: /blog|
			if i-offset == len(node.prefix) {
				return node.handle
			}

			// node: /blog|feed
			// path: /blog|
			return nil
		}

		// The node we just checked is entirely included in our path.
		// node: /|
		// path: /|blog
		if i-offset == len(node.prefix) {
			child := node.children[path[i]]

			if child != nil {
				offset = i
				node = child
				goto next
			}

			return nil
		}

		// We got a conflict.
		// node: /b|ag
		// path: /b|riefcase
		if path[i] != node.prefix[i-offset] {
			return nil
		}

	next:
		i++
	}
}

// prettyPrint
func (node *tree) prettyPrint(writer io.Writer) {
	fmt.Fprintf(writer, "%s (%d) [%t]\n", color.CyanString(node.prefix), len(node.children), node.handle != nil)

	for _, child := range node.children {
		fmt.Fprint(writer, "|_ ")
		child.prettyPrint(writer)
	}
}
