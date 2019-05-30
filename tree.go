package aero

import (
	"fmt"
	"io"
	"strings"

	"github.com/akyoto/color"
)

const (
	separator = '/'
	parameter = ':'
	wildcard  = '*'
)

// dataType specifies which type of data we are going to save for each node.
type dataType = interface{}

// tree represents a radix tree.
type tree struct {
	prefix    string
	data      dataType
	children  [256]*tree
	parameter *tree
	wildcard  *tree
	kind      byte
}

// add adds a new element to the tree.
func (node *tree) add(path string, data dataType) {
	// Search tree for equal parts until we can no longer proceed
	i := 0
	offset := 0

	for {
		if i == len(path) {
			// The path already exists.
			// node: /blog|
			// path: /blog|
			if i-offset == len(node.prefix) {
				node.data = data
				return
			}

			// The path ended but the node prefix is longer.
			// node: /blog|feed
			// path: /blog|
			node.split(i-offset, "", data)
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
				node.data = data
				return
			}

			// node: /user|
			// path: /user|/:userid
			if strings.Contains(path[i:], ":") || strings.Contains(path[i:], "*") {
				parameterKind := byte(0)
				start := i

				handleParameters := func() {
					// What was the parameter that we just scanned, if there were any?
					switch parameterKind {
					case parameter:
						name := path[start+1 : i]
						fmt.Println("PARAMETER NAME", name)
						node.parameter = &tree{
							prefix: name,
							kind:   parameter,
							data:   data,
						}
						node = node.parameter

					case wildcard:
						name := path[start+1 : i]
						fmt.Println("WILDCARD NAME", name)
						node.wildcard = &tree{
							prefix: name,
							kind:   wildcard,
							data:   data,
						}
						node = node.wildcard
					}
				}

				for i < len(path) {
					switch path[i] {
					case parameter, wildcard:
						parameterKind = path[i]

						child := &tree{
							prefix: path[start:i],
						}

						node.children[path[start]] = child
						node = child
						start = i

					case separator:
						handleParameters()
						parameterKind = 0
					}

					i++
				}

				// Trailing parameters
				handleParameters()

				return
			}

			// Otherwise, add a new child with the remaining string.
			node.children[path[i]] = &tree{
				prefix: path[i:],
				data:   data,
			}

			return
		}

		// We got a conflict.
		// node: /b|ag
		// path: /b|riefcase
		if path[i] != node.prefix[i-offset] {
			node.split(i-offset, path[i:], data)
			return
		}

	next:
		i++
	}
}

// split splits the node at the given index and inserts
// a new child node with the given path and data.
// If path is empty, it will not create another child node
// and instead assign the data directly to the node.
func (node *tree) split(index int, path string, data dataType) {
	// Create split node with the remaining string
	splitNode := &tree{
		prefix:   node.prefix[index:],
		data:     node.data,
		children: node.children,
	}

	// Cut the existing node
	node.prefix = node.prefix[:index]

	// If the path is empty, it means we don't create a 2nd child node.
	// Just assign the data for the existing node and store a single child node.
	if path == "" {
		node.data = data
		node.children[splitNode.prefix[0]] = splitNode
		return
	}

	// Create new node with the remaining string in the path
	newNode := &tree{
		prefix: path,
		data:   data,
	}

	// The existing data must be removed
	node.data = nil

	// Assign new child nodes
	node.children = [256]*tree{}
	node.children[splitNode.prefix[0]] = splitNode
	node.children[newNode.prefix[0]] = newNode
}

// find returns the data for the given path, if available.
func (node *tree) find(path string) dataType {
	// Search tree for equal parts until we can no longer proceed
	i := 0
	offset := 0

	for {
	beginning:
		// We reached the end.
		if i == len(path) {
			// node: /blog|
			// path: /blog|
			if node.kind != 0 || i-offset == len(node.prefix) {
				return node.data
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

			// node: /|:id
			// path: /|blog
			if node.parameter != nil {
				start := i

				for i < len(path) && path[i] != separator {
					i++
				}

				fmt.Printf("PARAMETER %s IS %s\n", node.parameter.prefix, path[start:i])

				node = node.parameter
				offset = i
				goto beginning
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

// PrettyPrint prints a human-readable form of the tree to the given writer.
func (node *tree) PrettyPrint(writer io.Writer) {
	node.prettyPrint(writer, -1)
}

// prettyPrint
func (node *tree) prettyPrint(writer io.Writer, level int) {
	prefix := ""

	if level >= 0 {
		prefix = strings.Repeat("  ", level) + "|_ "
	}

	colorFunc := color.CyanString

	switch node.kind {
	case ':':
		prefix += ":"
		colorFunc = color.YellowString
	case '*':
		prefix += "*"
		colorFunc = color.GreenString
	}

	fmt.Fprintf(writer, "%s%s [%t]\n", prefix, colorFunc(node.prefix), node.data != nil)

	for _, child := range node.children {
		if child == nil {
			continue
		}

		child.prettyPrint(writer, level+1)
	}

	if node.parameter != nil {
		node.parameter.prettyPrint(writer, level+1)
	}

	if node.wildcard != nil {
		node.wildcard.prettyPrint(writer, level+1)
	}
}
