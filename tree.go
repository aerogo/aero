package aero

import (
	"fmt"
	"io"
	"strings"

	"github.com/akyoto/color"
)

// node types
const (
	separator = '/'
	parameter = ':'
	wildcard  = '*'
)

// controlFlow tells the main loop what it should do next.
type controlFlow int

// controlFlow values.
const (
	controlStop  controlFlow = 0
	controlBegin controlFlow = 1
	controlNext  controlFlow = 2
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
	begin:
		switch node.kind {
		case parameter:
			// This only occurs when the same parameter based route is added twice.
			// node: /post/:id|
			// path: /post/:id|
			if i == len(path) {
				node.data = data
				return
			}

			// When we hit a separator, we'll search for a fitting child.
			if path[i] == separator {
				var control controlFlow
				node, offset, control = node.end(path, data, i, offset)

				switch control {
				case controlStop:
					return
				case controlBegin:
					goto begin
				case controlNext:
					goto next
				}
			}

		default:
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
				var control controlFlow
				node, offset, control = node.end(path, data, i, offset)

				switch control {
				case controlStop:
					return
				case controlBegin:
					goto begin
				case controlNext:
					goto next
				}
			}

			// We got a conflict.
			// node: /b|ag
			// path: /b|riefcase
			if path[i] != node.prefix[i-offset] {
				node.split(i-offset, path[i:], data)
				return
			}
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
	splitNode := node.clone(node.prefix[index:])

	/// The existing data must be removed
	node.reset(node.prefix[:index])

	// If the path is empty, it means we don't create a 2nd child node.
	// Just assign the data for the existing node and store a single child node.
	if path == "" {
		node.data = data
		node.children[splitNode.prefix[0]] = splitNode
		return
	}

	node.children[splitNode.prefix[0]] = splitNode

	// Create new nodes with the remaining path
	node.append(path, data)
}

// clone clones the node with a new prefix.
func (node *tree) clone(prefix string) *tree {
	return &tree{
		prefix:    prefix,
		data:      node.data,
		children:  node.children,
		parameter: node.parameter,
		wildcard:  node.wildcard,
		kind:      node.kind,
	}
}

// reset resets the existing node data.
func (node *tree) reset(prefix string) {
	node.prefix = prefix
	node.data = nil
	node.parameter = nil
	node.wildcard = nil
	node.kind = 0
	node.children = [256]*tree{}
}

// addTrailingSlash adds a trailing slash with the same data.
func (node *tree) addTrailingSlash(data dataType) {
	if strings.HasSuffix(node.prefix, "/") || node.children[separator] != nil || node.kind == wildcard {
		return
	}

	node.children[separator] = &tree{
		prefix: "/",
		data:   data,
	}
}

// append appends the given path to the tree.
func (node *tree) append(path string, data dataType) {
	// At this point, all we know is that somewhere
	// in the remaining string we have parameters.
	// node: /user|
	// path: /user|/:userid
	for {
		if path == "" {
			node.data = data
			return
		}

		paramStart := strings.IndexByte(path, parameter)

		if paramStart == -1 {
			paramStart = strings.IndexByte(path, wildcard)
		}

		// If it's a static route we are adding,
		// just add the remainder as a normal node.
		if paramStart == -1 {
			child := &tree{
				prefix: path,
				data:   data,
			}

			node.children[path[0]] = child
			child.addTrailingSlash(data)
			return
		}

		// If we're directly in front of a parameter,
		// add a parameter node.
		if paramStart == 0 {
			paramEnd := strings.IndexByte(path, separator)

			if paramEnd == -1 {
				paramEnd = len(path)
			}

			child := &tree{
				prefix: path[1:paramEnd],
				kind:   path[paramStart],
			}

			switch child.kind {
			case parameter:
				child.addTrailingSlash(data)
				node.parameter = child
				node = child
				path = path[paramEnd:]
				continue

			case wildcard:
				child.data = data
				node.wildcard = child
				return
			}
		}

		// We know there's a parameter, but not directly at the start.
		// Add a normal node with the path before the parameter start.
		child := &tree{
			prefix: path[:paramStart],
		}

		// Allow trailing slashes to return
		// the same content as their parent node.
		if child.prefix == "/" {
			child.data = node.data
		}

		node.children[path[0]] = child
		node = child
		path = path[paramStart:]
	}
}

// end is called when the node was fully parsed
// and needs to decide the next control flow.
func (node *tree) end(path string, data dataType, i int, offset int) (*tree, int, controlFlow) {
	child := node.children[path[i]]

	if child != nil {
		node = child
		offset = i
		return node, offset, controlNext
	}

	// No fitting children found, does this node even contain a prefix yet?
	// If no prefix is set, this is the starting node.
	if node.prefix == "" {
		node.prefix = path
		node.data = data
		return node, offset, controlStop
	}

	// node: /user/|:id
	// path: /user/|:id/profile
	if node.parameter != nil {
		node = node.parameter
		offset = i
		return node, offset, controlBegin
	}

	node.append(path[i:], data)
	return node, offset, controlStop
}

// find returns the data for the given path, if available.
func (node *tree) find(path string) dataType {
	var (
		i                  int
		offset             int
		lastWildcardOffset int
		lastWildcard       *tree
	)

	// Search tree for equal parts until we can no longer proceed
	for {
	begin:
		switch node.kind {
		case parameter:
			if i == len(path) {
				// fmt.Printf("PARAMETER %s IS %s\n", node.prefix, path[offset:i])
				return node.data
			}

			if path[i] == separator {
				// fmt.Printf("PARAMETER %s IS %s\n", node.prefix, path[offset:i])
				node = node.children[separator]
				offset = i
				goto next
			}

		default:
			// We reached the end.
			if i == len(path) {
				// node: /blog|
				// path: /blog|
				if i-offset == len(node.prefix) {
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
				if node.wildcard != nil {
					lastWildcard = node.wildcard
					lastWildcardOffset = i
				}

				child := node.children[path[i]]

				if child != nil {
					node = child
					offset = i
					goto next
				}

				// node: /|:id
				// path: /|blog
				if node.parameter != nil {
					node = node.parameter
					offset = i
					goto begin
				}

				// node: /|*any
				// path: /|image.png
				if node.wildcard != nil {
					fmt.Printf("WILCARD PARAMETER %s IS %s\n", node.wildcard.prefix, path[i:])
					return node.wildcard.data
				}

				return nil
			}

			// We got a conflict.
			// node: /b|ag
			// path: /b|riefcase
			if path[i] != node.prefix[i-offset] {
				if lastWildcard != nil {
					fmt.Printf("WILCARD PARAMETER %s IS %s\n", lastWildcard.prefix, path[lastWildcardOffset:])
					return lastWildcard.data
				}

				return nil
			}
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
