package chef

import (
	"path"
)

type (
	// Group represents a new routing group instance
	Group struct {
		prefix      string
		router      *Router
		middlewares []Handler
	}
)

// NewGroup returns a new routing group instance
func NewGroup(prefix string, router *Router) Group {
	g := Group{
		prefix: prefix,
		router: router,
	}

	return g
}

func (g *Group) add(method, p string, h Handler) {
	p = path.Clean(g.prefix + p)
	g.router.add(method, p, h, g.middlewares)
}

// Use adds middleware to the group chain.
func (g *Group) Use(middlewares ...Handler) {
	g.middlewares = append(g.middlewares, middlewares...)
}

// GET registers a new GET route for a path with matching handler in the router
func (g *Group) GET(path string, h Handler) {
	g.add("GET", path, h)
}

// POST registers a new POST route for a path with matching handler in the router
func (g *Group) POST(path string, h Handler) {
	g.add("POST", path, h)
}

// PUT registers a new PUT route for a path with matching handler in the router
func (g *Group) PUT(path string, h Handler) {
	g.add("PUT", path, h)
}

// PATCH registers a new PATCH route for a path with matching handler in the router
func (g *Group) PATCH(path string, h Handler) {
	g.add("PATCH", path, h)
}

// DELETE registers a new DELETE route for a path with matching handler in the router
func (g *Group) DELETE(path string, h Handler) {
	g.add("DELETE", path, h)
}

// CONNECT registers a new CONNECT route for a path with matching handler in the router
func (g *Group) CONNECT(path string, h Handler) {
	g.add("CONNECT", path, h)
}

// TRACE registers a new TRACE route for a path with matching handler in the router
func (g *Group) TRACE(path string, h Handler) {
	g.add("TRACE", path, h)
}

// OPTIONS registers a new OPTIONS route for a path with matching handler in the router
func (g *Group) OPTIONS(path string, h Handler) {
	g.add("OPTIONS", path, h)
}
