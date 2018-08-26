package chef

import (
	"net/http"
	"sync"
)

type (
	// Handler represents a function to handle HTTP requests
	Handler func(Context)

	route struct {
		Method string
		Path   string
		Name   string
	}

	// Router represents a new router instance
	Router struct {
		tree        *node
		pool        sync.Pool
		routes      map[string]*route
		middlewares []Handler
		after       []Handler
		config      *Config
		maxParam    *int
	}
)

// Error Handlers
var (
	NotFoundHandler = func(c Context) {
		c.SetStatusCode(http.StatusNotFound)
		c.WriteString("Error 404: not found")
	}

	MethodNotAllowedHandler = func(c Context) {
		c.SetStatusCode(http.StatusMethodNotAllowed)
		c.WriteString("method not allowed")
	}
)

// NewRouter returns a router instance
func NewRouter(config *Config) *Router {
	r := &Router{
		tree: &node{
			methodHandler: new(methodHandler),
		},
		routes:   map[string]*route{},
		config:   config,
		maxParam: new(int),
	}
	r.pool.New = func() interface{} {
		return NewContext(nil, nil, r.maxParam)
	}

	return r
}

// Add registers a new route for method and path with matching handler.
func (r *Router) add(method, path string, h Handler, hs []Handler) {
	// Validate path
	if path == "" {
		panic("chef: path cannot be empty")
	}
	if path[0] != '/' {
		path = "/" + path
	}
	pnames := []string{} // Param names
	ppath := path        // Pristine path

	handlers := r.middlewares
	if hs != nil {
		handlers = append(handlers, hs...)
	}
	handlers = append(handlers, h)
	handlers = append(handlers, r.after...)

	for i, l := 0, len(path); i < l; i++ {
		if path[i] == ':' {
			j := i + 1

			r.insert(method, path[:i], nil, skind, "", nil)
			for ; i < l && path[i] != '/'; i++ {
			}

			pnames = append(pnames, path[j:i])
			path = path[:j] + path[i:]
			i, l = j, len(path)

			if i == l {
				r.insert(method, path[:i], handlers, pkind, ppath, pnames)
				return
			}
			r.insert(method, path[:i], nil, pkind, ppath, pnames)
		} else if path[i] == '*' {
			r.insert(method, path[:i], nil, skind, "", nil)
			pnames = append(pnames, "*")
			r.insert(method, path[:i+1], handlers, akind, ppath, pnames)
			return
		}
	}

	r.insert(method, path, handlers, skind, ppath, pnames)
}

func (r *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	ctx := r.pool.Get().(*context)
	defer r.pool.Put(ctx)
	ctx.reset(req, res, r.config)

	method := req.Method
	path := req.URL.RawPath
	if path == "" {
		path = req.URL.Path
	}

	r.Find(method, path, ctx)

	ctx.Next()
}
