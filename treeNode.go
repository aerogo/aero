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

// treeNode represents a radix tree node.
type treeNode struct {
	prefix     string
	startIndex uint8
	endIndex   uint8
	indices    []uint8
	children   []*treeNode
	data       dataType
	parameter  *treeNode
	wildcard   *treeNode
	kind       byte
}

// split splits the node at the given index and inserts
// a new child node with the given path and data.
// If path is empty, it will not create another child node
// and instead assign the data directly to the node.
func (node *treeNode) split(index int, path string, data dataType) {
	// Create split node with the remaining string
	splitNode := node.clone(node.prefix[index:])

	// The existing data must be removed
	node.reset(node.prefix[:index])

	// If the path is empty, it means we don't create a 2nd child node.
	// Just assign the data for the existing node and store a single child node.
	if path == "" {
		node.data = data
		node.addChild(splitNode)
		return
	}

	node.addChild(splitNode)

	// Create new nodes with the remaining path
	node.append(path, data)
}

// clone clones the node with a new prefix.
func (node *treeNode) clone(prefix string) *treeNode {
	return &treeNode{
		prefix:     prefix,
		data:       node.data,
		indices:    node.indices,
		startIndex: node.startIndex,
		endIndex:   node.endIndex,
		children:   node.children,
		parameter:  node.parameter,
		wildcard:   node.wildcard,
		kind:       node.kind,
	}
}

// reset resets the existing node data.
func (node *treeNode) reset(prefix string) {
	node.prefix = prefix
	node.data = nil
	node.parameter = nil
	node.wildcard = nil
	node.kind = 0
	node.startIndex = 0
	node.endIndex = 0
	node.indices = nil
	node.children = nil
}

// addChild adds a child tree.
func (node *treeNode) addChild(child *treeNode) {
	if len(node.children) == 0 {
		node.children = append(node.children, nil)
	}

	firstChar := child.prefix[0]

	switch {
	case node.startIndex == 0:
		node.startIndex = firstChar
		node.indices = []uint8{0}
		node.endIndex = node.startIndex + uint8(len(node.indices))

	case firstChar < node.startIndex:
		diff := node.startIndex - firstChar
		newIndices := make([]uint8, diff+uint8(len(node.indices)))
		copy(newIndices[diff:], node.indices)
		node.startIndex = firstChar
		node.indices = newIndices
		node.endIndex = node.startIndex + uint8(len(node.indices))

	case firstChar >= node.endIndex:
		diff := firstChar - node.endIndex + 1
		newIndices := make([]uint8, diff+uint8(len(node.indices)))
		copy(newIndices, node.indices)
		node.indices = newIndices
		node.endIndex = node.startIndex + uint8(len(node.indices))
	}

	index := node.indices[firstChar-node.startIndex]

	if index == 0 {
		node.indices[firstChar-node.startIndex] = uint8(len(node.children))
		node.children = append(node.children, child)
		return
	}

	node.children[index] = child
}

// addTrailingSlash adds a trailing slash with the same data.
func (node *treeNode) addTrailingSlash(data dataType) {
	if strings.HasSuffix(node.prefix, "/") || node.kind == wildcard || (separator >= node.startIndex && separator < node.endIndex && node.indices[separator-node.startIndex] != 0) {
		return
	}

	node.addChild(&treeNode{
		prefix: "/",
		data:   data,
	})
}

// append appends the given path to the tree.
func (node *treeNode) append(path string, data dataType) {
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
			// If the node itself doesn't have a prefix (root node),
			// don't add a child and use the node itself.
			if node.prefix == "" {
				node.prefix = path
				node.data = data
				return
			}

			child := &treeNode{
				prefix: path,
				data:   data,
			}

			node.addChild(child)
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

			child := &treeNode{
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

		// If the node itself doesn't have a prefix (root node),
		// don't add a child and use the node itself.
		if node.prefix == "" {
			node.prefix = path[:paramStart]
			path = path[paramStart:]
			continue
		}

		// Add a normal node with the path before the parameter start.
		child := &treeNode{
			prefix: path[:paramStart],
		}

		// Allow trailing slashes to return
		// the same content as their parent node.
		if child.prefix == "/" {
			child.data = node.data
		}

		node.addChild(child)
		node = child
		path = path[paramStart:]
	}
}

// end is called when the node was fully parsed
// and needs to decide the next control flow.
func (node *treeNode) end(path string, data dataType, i int, offset int) (*treeNode, int, controlFlow) {
	char := path[i]

	if char >= node.startIndex && char < node.endIndex {
		index := node.indices[char-node.startIndex]

		if index != 0 {
			node = node.children[index]
			offset = i
			return node, offset, controlNext
		}
	}

	// No fitting children found, does this node even contain a prefix yet?
	// If no prefix is set, this is the starting node.
	if node.prefix == "" {
		node.append(path[i:], data)
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

// each traverses the tree and calls the given function on every node.
func (node *treeNode) each(callback func(*treeNode)) {
	callback(node)

	for _, child := range node.children {
		if child == nil {
			continue
		}

		child.each(callback)
	}

	if node.parameter != nil {
		node.parameter.each(callback)
	}

	if node.wildcard != nil {
		node.wildcard.each(callback)
	}
}

// PrettyPrint prints a human-readable form of the tree to the given writer.
func (node *treeNode) PrettyPrint(writer io.Writer) {
	node.prettyPrint(writer, -1)
}

// prettyPrint
func (node *treeNode) prettyPrint(writer io.Writer, level int) {
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
