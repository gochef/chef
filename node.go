package chef

import "fmt"

type (
	kind uint8
	node struct {
		kind          kind
		label         byte
		prefix        string
		parent        *node
		children      children
		ppath         string
		pnames        []string
		methodHandler *methodHandler
	}
	children      []*node
	methodHandler struct {
		connect []Handler
		delete  []Handler
		get     []Handler
		head    []Handler
		options []Handler
		patch   []Handler
		post    []Handler
		put     []Handler
		trace   []Handler
	}
)

const (
	skind kind = iota
	pkind
	akind
)

var (
	methods = [...]string{
		CONNECT,
		DELETE,
		GET,
		HEAD,
		OPTIONS,
		PATCH,
		POST,
		PUT,
		TRACE,
	}
)

func newNode(t kind, pre string, p *node, c children, mh *methodHandler, ppath string, pnames []string) *node {
	return &node{
		kind:          t,
		label:         pre[0],
		prefix:        pre,
		parent:        p,
		children:      c,
		ppath:         ppath,
		pnames:        pnames,
		methodHandler: mh,
	}
}

func (n *node) addChild(c *node) {
	n.children = append(n.children, c)
}

func (n *node) findChild(l byte, t kind) *node {
	for _, c := range n.children {
		if c.label == l && c.kind == t {
			return c
		}
	}
	return nil
}

func (n *node) findChildWithLabel(l byte) *node {
	for _, c := range n.children {
		if c.label == l {
			return c
		}
	}
	return nil
}

func (n *node) findChildByKind(t kind) *node {
	for _, c := range n.children {
		if c.kind == t {
			return c
		}
	}
	return nil
}

func (n *node) addHandler(method string, h []Handler) {
	switch method {
	case GET:
		n.methodHandler.get = h
	case POST:
		n.methodHandler.post = h
	case PUT:
		n.methodHandler.put = h
	case DELETE:
		n.methodHandler.delete = h
	case PATCH:
		n.methodHandler.patch = h
	case OPTIONS:
		n.methodHandler.options = h
	case HEAD:
		n.methodHandler.head = h
	case CONNECT:
		n.methodHandler.connect = h
	case TRACE:
		n.methodHandler.trace = h
	}
}

func (n *node) findHandler(method string) []Handler {
	switch method {
	case GET:
		return n.methodHandler.get
	case POST:
		return n.methodHandler.post
	case PUT:
		return n.methodHandler.put
	case DELETE:
		return n.methodHandler.delete
	case PATCH:
		return n.methodHandler.patch
	case OPTIONS:
		return n.methodHandler.options
	case HEAD:
		return n.methodHandler.head
	case CONNECT:
		return n.methodHandler.connect
	case TRACE:
		return n.methodHandler.trace
	default:
		return nil
	}
}

func (n *node) checkMethodNotAllowed() []Handler {
	for _, m := range methods {
		if h := n.findHandler(m); h != nil {
			hs := []Handler{
				MethodNotAllowedHandler,
			}
			return hs
		}
	}
	hs := []Handler{
		NotFoundHandler,
	}
	return hs
}

func (n *node) insert(method, path string, h []Handler, t kind, ppath string, pnames []string) {
	if n == nil {
		panic("chef: invalid method")
	}
	search := path

	for {
		sl := len(search)
		pl := len(n.prefix)
		l := 0

		// LCP
		max := pl
		if sl < max {
			max = sl
		}
		for ; l < max && search[l] == n.prefix[l]; l++ {
		}

		if l == 0 {
			// At root node
			n.label = search[0]
			n.prefix = search
			if h != nil {
				n.kind = t
				n.addHandler(method, h)
				n.ppath = ppath
				n.pnames = pnames
			}
		} else if l < pl {
			// Split node
			nNode := newNode(n.kind, n.prefix[l:], n, n.children, n.methodHandler, n.ppath, n.pnames)

			// Reset parent node
			n.kind = skind
			n.label = n.prefix[0]
			n.prefix = n.prefix[:l]
			n.children = nil
			n.methodHandler = new(methodHandler)
			n.ppath = ""
			n.pnames = nil

			n.addChild(nNode)

			if l == sl {
				// At parent node
				n.kind = t
				n.addHandler(method, h)
				n.ppath = ppath
				n.pnames = pnames
			} else {
				// Create child node
				nNode = newNode(t, search[l:], n, nil, new(methodHandler), ppath, pnames)
				nNode.addHandler(method, h)
				n.addChild(nNode)
			}
		} else if l < sl {
			search = search[l:]
			c := n.findChildWithLabel(search[0])
			if c != nil {
				// Go deeper
				n = c
				continue
			}
			// Create child node
			nNode := newNode(t, search, n, nil, new(methodHandler), ppath, pnames)
			nNode.addHandler(method, h)
			n.addChild(nNode)
		} else {
			// Node already exists
			if h != nil {
				n.addHandler(method, h)
				n.ppath = ppath
				if len(n.pnames) == 0 { // Issue #729
					n.pnames = pnames
				}
			}
		}
		return
	}
}

// Find lookup a handler registered for method and path. It also parses URL for path
// parameters and load them into context.
//
// For performance:
//
// - Get context from `Echo#AcquireContext()`
// - Reset it `Context#Reset()`
// - Return it `Echo#ReleaseContext()`.
func (n *node) find(method, path string, c Context) {
	ctx := c.(*context)
	ctx.path = path

	var (
		search  = path
		child   *node         // Child node
		nc      int           // Param counter
		nk      kind          // Next kind
		nn      *node         // Next node
		ns      string        // Next search
		pvalues = ctx.pvalues // Use the internal slice so the interface can keep the illusion of a dynamic slice
	)

	// Search order static > param > any
	for {
		if search == "" {
			goto End
		}

		pl := 0 // Prefix length
		l := 0  // LCP length

		if n.label != ':' {
			sl := len(search)
			pl = len(n.prefix)

			// LCP
			max := pl
			if sl < max {
				max = sl
			}
			for ; l < max && search[l] == n.prefix[l]; l++ {
			}
		}

		if l == pl {
			// Continue search
			search = search[l:]
		} else {
			n = nn
			search = ns
			if nk == pkind {
				goto Param
			} else if nk == akind {
				goto Any
			}
			// Not found
			return
		}

		if search == "" {
			goto End
		}

		// Static node
		if child = n.findChild(search[0], skind); child != nil {
			// Save next
			if n.prefix[len(n.prefix)-1] == '/' { // Issue #623
				nk = pkind
				nn = n
				ns = search
			}
			n = child
			continue
		}

		// Param node
	Param:
		if child = n.findChildByKind(pkind); child != nil {
			// Issue #378
			if len(pvalues) == nc {
				continue
			}

			// Save next
			if n.prefix[len(n.prefix)-1] == '/' { // Issue #623
				nk = akind
				nn = n
				ns = search
			}

			n = child
			i, l := 0, len(search)
			for ; i < l && search[i] != '/'; i++ {
			}
			pvalues[nc] = search[:i]
			nc++
			search = search[i:]
			continue
		}

		// Any node
	Any:
		if n = n.findChildByKind(akind); n == nil {
			if nn != nil {
				n = nn
				nn = n.parent // Next (Issue #954)
				search = ns
				if nk == pkind {
					goto Param
				} else if nk == akind {
					goto Any
				}
			}
			// Not found
			return
		}

		if len(pvalues) > 0 {
			fmt.Println("ff")
			pvalues[len(n.pnames)-1] = search
		}

		pnamesLength := len(n.pnames)
		if len(pvalues) == pnamesLength {
			pvalues[pnamesLength-1] = search
		}

		/**pnameLength := len(n.pnames) - 1
		if len(pvalues) >= pnameLength+1 {
			pvalues[pnameLength] = search
		}
		//pvalues[len(n.pnames)-1] = search**/
		goto End
	}

End:
	ctx.SetHandlers(n.findHandler(method))
	ctx.path = n.ppath
	ctx.pnames = n.pnames

	// NOTE: Slow zone...
	if ctx.GetHandlers() == nil {

		ctx.SetHandlers(n.checkMethodNotAllowed())

		// Dig further for any, might have an empty value for *, e.g.
		// serving a directory. Issue #207.
		if n = n.findChildByKind(akind); n == nil {
			return
		}
		if h := n.findHandler(method); h != nil {
			ctx.SetHandlers(h)
		} else {
			ctx.SetHandlers(n.checkMethodNotAllowed())
		}
		ctx.path = n.ppath
		ctx.pnames = n.pnames
		pvalues[len(n.pnames)-1] = ""
	}

	return
}
