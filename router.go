package pine

import "strings"

// node is a single node in the radix tree.
// Children are looked up by the first byte of their path segment.
// Priority for matching: static children first, then paramChild, then wildcard.
type node struct {
	path      string             // static path segment stored at this node
	children  map[byte]*node     // static children keyed by first byte of segment
	paramChild *node             // handles :param segments
	wildcard  *node              // handles /* catch-all
	handlers  map[string][]Handler // method → handler slice (at leaf nodes)
	paramNames []string          // param names accumulated from root to this node
}

// Router is a radix-tree HTTP router.
// Insert registers a route; Search looks one up.
type Router struct {
	root *node
}

func newRouter() *Router {
	return &Router{root: &node{}}
}

// Insert registers handlers for a given HTTP method and path.
// Segments starting with ':' are treated as named parameters.
// A trailing '/*' is treated as a catch-all wildcard.
func (r *Router) Insert(method, path string, handlers []Handler) {
	if path == "" {
		path = "/"
	}
	// Trim leading slash so we work with the suffix only.
	if path[0] == '/' {
		path = path[1:]
	}
	r.root.insert(method, path, handlers, nil)
}

// Search finds the handlers registered for method+path.
// Returns (handlers, params, found).
func (r *Router) Search(method, path string) ([]Handler, map[string]string, bool) {
	if path == "" {
		path = "/"
	}
	if path[0] == '/' {
		path = path[1:]
	}
	return r.root.search(method, path, nil)
}

// SearchAnyMethod returns handlers for path regardless of method.
// Used for OPTIONS / CORS preflight fallback.
func (r *Router) SearchAnyMethod(path string) ([]Handler, map[string]string, bool) {
	if path == "" {
		path = "/"
	}
	if path[0] == '/' {
		path = path[1:]
	}
	return r.root.searchAnyMethod(path, nil)
}

// insert walks the tree and registers handlers at the terminal node.
func (n *node) insert(method, path string, handlers []Handler, paramNames []string) {
	// Reached the end of the path — this is the leaf.
	if path == "" {
		if n.handlers == nil {
			n.handlers = make(map[string][]Handler)
		}
		n.handlers[method] = handlers
		n.paramNames = paramNames
		return
	}

	// Consume the next segment (up to the next '/' or end of string).
	seg, rest := nextSegment(path)

	if seg == "" {
		// Consecutive or trailing slash — skip and continue.
		n.insert(method, rest, handlers, paramNames)
		return
	}

	firstByte := seg[0]

	switch firstByte {
	case ':':
		// Parametric segment.
		paramName := seg[1:]
		if n.paramChild == nil {
			n.paramChild = &node{}
		}
		n.paramChild.insert(method, rest, handlers, append(paramNames, paramName))

	case '*':
		// Catch-all wildcard. Remaining path captured as one param.
		paramName := seg[1:]
		if paramName == "" {
			paramName = "wildcard"
		}
		if n.wildcard == nil {
			n.wildcard = &node{}
		}
		if n.wildcard.handlers == nil {
			n.wildcard.handlers = make(map[string][]Handler)
		}
		n.wildcard.handlers[method] = handlers
		n.wildcard.paramNames = append(paramNames, paramName)

	default:
		// Static segment.
		if n.children == nil {
			n.children = make(map[byte]*node)
		}
		child, ok := n.children[firstByte]
		if !ok {
			child = &node{path: seg}
			n.children[firstByte] = child
			child.insert(method, rest, handlers, paramNames)
			return
		}

		// Radix compression: find longest common prefix between child.path and seg.
		lcp := longestCommonPrefix(child.path, seg)

		if lcp == len(child.path) {
			// Existing child covers our prefix — consume it and keep going.
			child.insert(method, strings.TrimPrefix(seg, child.path)+segSep(rest), handlers, paramNames)
			return
		}

		// Need to split the existing child at lcp.
		split := &node{
			path:     child.path[:lcp],
			children: map[byte]*node{child.path[lcp]: child},
		}
		child.path = child.path[lcp:]
		n.children[firstByte] = split

		// Now insert the new route into the split node.
		remainingNew := seg[lcp:]
		if remainingNew == "" {
			split.insert(method, rest, handlers, paramNames)
		} else {
			split.insert(method, remainingNew+segSep(rest), handlers, paramNames)
		}
	}
}

// search walks the tree and returns matching handlers + captured params.
func (n *node) search(method, path string, captured []string) ([]Handler, map[string]string, bool) {
	// End of path — check for a handler at this leaf.
	if path == "" {
		if h, ok := n.handlers[method]; ok {
			return h, buildParams(n.paramNames, captured), true
		}
		return nil, nil, false
	}

	seg, rest := nextSegment(path)

	if seg == "" {
		// An empty segment means consecutive slashes in the request path.
		// No registered route can match this, so return not-found.
		return nil, nil, false
	}

	firstByte := seg[0]

	// 1. Try static children first (highest priority).
	if n.children != nil {
		if child, ok := n.children[firstByte]; ok {
			if strings.HasPrefix(seg, child.path) {
				remainder := seg[len(child.path):]
				if remainder == "" {
					if h, pm, found := child.search(method, rest, captured); found {
						return h, pm, true
					}
				} else {
					// The segment is longer than child.path — try descending.
					if h, pm, found := child.search(method, remainder+segSep(rest), captured); found {
						return h, pm, true
					}
				}
			}
		}
	}

	// 2. Try parametric child.
	if n.paramChild != nil && len(seg) > 0 {
		if h, pm, found := n.paramChild.search(method, rest, append(captured, seg)); found {
			return h, pm, true
		}
	}

	// 3. Try wildcard (catch-all): capture everything remaining.
	if n.wildcard != nil {
		if h, ok := n.wildcard.handlers[method]; ok {
			remaining := path
			return h, buildParams(n.wildcard.paramNames, append(captured, remaining)), true
		}
	}

	return nil, nil, false
}

// searchAnyMethod matches any registered method for OPTIONS/CORS fallback.
func (n *node) searchAnyMethod(path string, captured []string) ([]Handler, map[string]string, bool) {
	if path == "" {
		for _, h := range n.handlers {
			return h, buildParams(n.paramNames, captured), true
		}
		return nil, nil, false
	}

	seg, rest := nextSegment(path)
	if seg == "" {
		return nil, nil, false
	}

	firstByte := seg[0]

	if n.children != nil {
		if child, ok := n.children[firstByte]; ok && strings.HasPrefix(seg, child.path) {
			remainder := seg[len(child.path):]
			if remainder == "" {
				if h, pm, found := child.searchAnyMethod(rest, captured); found {
					return h, pm, true
				}
			} else {
				if h, pm, found := child.searchAnyMethod(remainder+segSep(rest), captured); found {
					return h, pm, true
				}
			}
		}
	}

	if n.paramChild != nil && len(seg) > 0 {
		if h, pm, found := n.paramChild.searchAnyMethod(rest, append(captured, seg)); found {
			return h, pm, true
		}
	}

	if n.wildcard != nil {
		for _, h := range n.wildcard.handlers {
			return h, buildParams(n.wildcard.paramNames, append(captured, path)), true
		}
	}

	return nil, nil, false
}

// nextSegment splits path at the first '/' returning (segment, remainder).
// The leading '/' in remainder is consumed.
func nextSegment(path string) (seg, rest string) {
	idx := strings.IndexByte(path, '/')
	if idx == -1 {
		return path, ""
	}
	return path[:idx], path[idx+1:]
}

// segSep returns "/" + rest when rest is non-empty, so segments can be
// rejoined correctly during split/insert operations.
func segSep(rest string) string {
	if rest == "" {
		return ""
	}
	return "/" + rest
}

// longestCommonPrefix returns the length of the shared prefix between a and b.
func longestCommonPrefix(a, b string) int {
	max := len(a)
	if len(b) < max {
		max = len(b)
	}
	for i := 0; i < max; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return max
}

// buildParams zips paramNames with captured values into a map.
func buildParams(names, values []string) map[string]string {
	if len(names) == 0 {
		return make(map[string]string)
	}
	params := make(map[string]string, len(names))
	for i, name := range names {
		if i < len(values) {
			params[name] = values[i]
		}
	}
	return params
}
