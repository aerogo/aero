package aero

import (
	"strings"
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
type dataType = Handler

// tree represents a radix tree.
type tree struct {
	root        treeNode
	static      map[string]dataType
	canBeStatic [2048]bool
}

// add adds a new element to the tree.
func (tree *tree) add(path string, data dataType) {
	if !strings.Contains(path, ":") && !strings.Contains(path, "*") {
		if tree.static == nil {
			tree.static = map[string]dataType{}
		}

		tree.static[path] = data
		tree.canBeStatic[len(path)] = true
		return
	}

	// Search tree for equal parts until we can no longer proceed
	i := 0
	offset := 0
	node := &tree.root

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

// find finds the data for the given path and assigns it to ctx.handler, if available.
func (tree *tree) find(path string, ctx *context) {
	if tree.canBeStatic[len(path)] {
		handler, found := tree.static[path]

		if found {
			ctx.handler = handler
			return
		}
	}

	var (
		i                  uint
		offset             uint
		lastWildcardOffset uint
		lastWildcard       *treeNode
		node               = &tree.root
	)

	// Search tree for equal parts until we can no longer proceed
	for {
		if node.kind == parameter {
			for {
				// We reached the end.
				if i == uint(len(path)) {
					ctx.addParameter(node.prefix, path[offset:i])
					ctx.handler = node.data
					return
				}

				// node: /:xxx|/:yyy
				// path: /blog|/post
				if path[i] == separator {
					ctx.addParameter(node.prefix, path[offset:i])
					index := node.indices[separator-node.startIndex]
					node = node.children[index]
					offset = i
					i++
					break
				}

				i++
			}

			continue
		}

		// We reached the end.
		if i == uint(len(path)) {
			// node: /blog|
			// path: /blog|
			if i-offset == uint(len(node.prefix)) {
				ctx.handler = node.data
				return
			}

			// node: /blog|feed
			// path: /blog|
			ctx.handler = nil
			return
		}

		// The node we just checked is entirely included in our path.
		// node: /|
		// path: /|blog
		if i-offset == uint(len(node.prefix)) {
			if node.wildcard != nil {
				lastWildcard = node.wildcard
				lastWildcardOffset = i
			}

			char := path[i]

			if char >= node.startIndex && char < node.endIndex {
				index := node.indices[char-node.startIndex]

				if index != 0 {
					node = node.children[index]
					offset = i
					i++
					continue
				}
			}

			// node: /|:id
			// path: /|blog
			if node.parameter != nil {
				node = node.parameter
				offset = i
				continue
			}

			// node: /|*any
			// path: /|image.png
			if node.wildcard != nil {
				ctx.addParameter(node.wildcard.prefix, path[i:])
				ctx.handler = node.wildcard.data
				return
			}

			ctx.handler = nil
			return
		}

		// We got a conflict.
		// node: /b|ag
		// path: /b|riefcase
		if path[i] != node.prefix[i-offset] {
			if lastWildcard != nil {
				ctx.addParameter(lastWildcard.prefix, path[lastWildcardOffset:])
				ctx.handler = lastWildcard.data
				return
			}

			ctx.handler = nil
			return
		}

		i++
	}
}

// bind binds all handlers to a new one provided by the callback.
func (tree *tree) bind(transform func(Handler) Handler) {
	tree.root.each(func(node *treeNode) {
		if node.data != nil {
			node.data = transform(node.data)
		}
	})

	for key, value := range tree.static {
		tree.static[key] = transform(value)
	}
}
