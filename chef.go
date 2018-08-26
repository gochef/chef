package chef

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/gochef/cache"
	"github.com/gochef/chef/utils"
	"github.com/gochef/session"
)

type (
	// Config represents a config instance
	Config struct {
		App struct {
			Name     string
			Static   string
			ViewPath string
			Port     string
			Env      string
		}
		Database struct {
			Driver      string
			Host        string
			Port        string
			Username    string
			Password    string
			Dbname      string
			AutoConnect bool
		}
		Fileserver struct {
			Use  bool
			Path string
			Dir  string
		}
		Cache   *cache.Config
		Session *session.Config
		Logger  *utils.LoggerConfig
	}

	// Data represents a map to store contextual data
	Data map[string]interface{}

	// Chef is the framework instance
	Chef struct {
		config *Config
		router *Router
		logger *utils.Logger
	}
)

// MIME types
const (
	MIMEApplicationJSON                  = "application/json"
	MIMEApplicationJSONCharsetUTF8       = MIMEApplicationJSON + "; " + charsetUTF8
	MIMEApplicationJavaScript            = "application/javascript"
	MIMEApplicationJavaScriptCharsetUTF8 = MIMEApplicationJavaScript + "; " + charsetUTF8
	MIMEApplicationXML                   = "application/xml"
	MIMEApplicationXMLCharsetUTF8        = MIMEApplicationXML + "; " + charsetUTF8
	MIMETextXML                          = "text/xml"
	MIMETextXMLCharsetUTF8               = MIMETextXML + "; " + charsetUTF8
	MIMEApplicationForm                  = "application/x-www-form-urlencoded"
	MIMEApplicationProtobuf              = "application/protobuf"
	MIMEApplicationMsgpack               = "application/msgpack"
	MIMETextHTML                         = "text/html"
	MIMETextHTMLCharsetUTF8              = MIMETextHTML + "; " + charsetUTF8
	MIMETextPlain                        = "text/plain"
	MIMETextPlainCharsetUTF8             = MIMETextPlain + "; " + charsetUTF8
	MIMEMultipartForm                    = "multipart/form-data"
	MIMEOctetStream                      = "application/octet-stream"
	MIMEApplicationAjax                  = "xmlhttprequest"
)

const (
	charsetUTF8 = "charset=UTF-8"
)

// Headers
const (
	HeaderAccept              = "Accept"
	HeaderAcceptEncoding      = "Accept-Encoding"
	HeaderAllow               = "Allow"
	HeaderAuthorization       = "Authorization"
	HeaderContentDisposition  = "Content-Disposition"
	HeaderContentEncoding     = "Content-Encoding"
	HeaderContentLength       = "Content-Length"
	HeaderContentType         = "Content-Type"
	HeaderCookie              = "Cookie"
	HeaderSetCookie           = "Set-Cookie"
	HeaderIfModifiedSince     = "If-Modified-Since"
	HeaderLastModified        = "Last-Modified"
	HeaderLocation            = "Location"
	HeaderUpgrade             = "Upgrade"
	HeaderVary                = "Vary"
	HeaderWWWAuthenticate     = "WWW-Authenticate"
	HeaderXForwardedFor       = "X-Forwarded-For"
	HeaderXForwardedProto     = "X-Forwarded-Proto"
	HeaderXForwardedProtocol  = "X-Forwarded-Protocol"
	HeaderXForwardedSsl       = "X-Forwarded-Ssl"
	HeaderXRequestedWith      = "X-Requested-With"
	HeaderXUrlScheme          = "X-Url-Scheme"
	HeaderXHTTPMethodOverride = "X-HTTP-Method-Override"
	HeaderXRealIP             = "X-Real-IP"
	HeaderXRequestID          = "X-Request-ID"
	HeaderServer              = "Server"
	HeaderOrigin              = "Origin"

	// Access control
	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"

	// Security
	HeaderStrictTransportSecurity = "Strict-Transport-Security"
	HeaderXContentTypeOptions     = "X-Content-Type-Options"
	HeaderXXSSProtection          = "X-XSS-Protection"
	HeaderXFrameOptions           = "X-Frame-Options"
	HeaderContentSecurityPolicy   = "Content-Security-Policy"
	HeaderXCSRFToken              = "X-CSRF-Token"
)

// HTTP methods
const (
	CONNECT = "CONNECT"
	DELETE  = "DELETE"
	GET     = "GET"
	HEAD    = "HEAD"
	OPTIONS = "OPTIONS"
	PATCH   = "PATCH"
	POST    = "POST"
	PUT     = "PUT"
	TRACE   = "TRACE"
)

var (
	defaultLogModules = []string{
		"chef",
		"chef.config",
		"chef.session",
		"chef.cache",
	}
)

// New returns an instance of the framework
func New() *Chef {
	c := &Chef{}

	// load and parse config
	c.loadConfig()

	// initialize logger
	c.config.Logger.Modules = append(defaultLogModules, c.config.Logger.Modules...)
	c.logger = utils.NewLogger(c.config.Logger)

	// start router
	c.router = NewRouter(c.config)

	// start fileserver
	if c.config.Fileserver.Use {
		c.startFileServer()
	}

	// Start session if configured to do so
	session.New(c.config.Session)

	return c
}

func (c *Chef) loadConfig() {
	if _, err := toml.DecodeFile("config.toml", &c.config); err != nil {
		panic("chef: Unable to load config: " + err.Error())
	}
}

// Logger returns a logger instance
func (c *Chef) Logger() *utils.Logger {
	return c.logger
}

// Config returns the application config
func (c *Chef) Config() *Config {
	return c.config
}

// Group returns a new routing group
func (c *Chef) Group(prefix string, cb func(Group)) {
	group := NewGroup(prefix, c.router)
	cb(group)
}

// Use registeres application-wide middlewares
func (c *Chef) Use(middlewares ...Handler) {
	c.router.middlewares = append(c.router.middlewares, middlewares...)
}

// After registers middlewares to be run after the main request handler
func (c *Chef) After(middlewares ...Handler) {
	c.router.after = append(c.router.after, middlewares...)
}

// GET registers a GET route for path with handler
func (c *Chef) GET(path string, h Handler) {
	c.router.add("GET", path, h, nil)
}

// POST registers a POST route for path with handler
func (c *Chef) POST(path string, h Handler) {
	c.router.add("POST", path, h, nil)
}

// PUT registers a PUT route for path with handler
func (c *Chef) PUT(path string, h Handler) {
	c.router.add("PUT", path, h, nil)
}

// PATCH registers a PATCH route for path with handler
func (c *Chef) PATCH(path string, h Handler) {
	c.router.add("PATCH", path, h, nil)
}

// DELETE registers a DELETE route for path with handler
func (c *Chef) DELETE(path string, h Handler) {
	c.router.add("DELETE", path, h, nil)
}

// CONNECT registers a CONNECT route for path with handler
func (c *Chef) CONNECT(path string, h Handler) {
	c.router.add("CONNECT", path, h, nil)
}

// TRACE registers a TRACE route for path with handler
func (c *Chef) TRACE(path string, h Handler) {
	c.router.add("TRACE", path, h, nil)
}

// OPTIONS registers a OPTIONS route for path with handler
func (c *Chef) OPTIONS(path string, h Handler) {
	c.router.add("OPTIONS", path, h, nil)
}

// All registers a new route for multiple HTTP methods and path with matching
// handler in the router with optional route-level middleware.
func (c *Chef) All(path string, handler Handler) {
	for _, m := range methods {
		c.router.add(m, path, handler, nil)
	}
}

// Some registers a new route for multiple HTTP methods and path with matching
// handler in the router with optional route-level middleware.
func (c *Chef) Some(mthds []string, path string, handler Handler) {
	for _, m := range mthds {
		c.router.add(m, path, handler, nil)
	}
}

func (c *Chef) startFileServer() {
	root := c.config.Fileserver.Dir
	path := c.config.Fileserver.Path

	workDir, _ := os.Getwd()
	filesDir := filepath.Join(workDir, root)
	dir := http.Dir(filesDir)

	fs := http.StripPrefix(path, http.FileServer(dir))
	if path != "/" && path[len(path)-1] != '/' {
		c.GET(path, func(c Context) {
			http.RedirectHandler(path+"/", 301).ServeHTTP(c.Response(), c.Request())
		})
		path += "/"
	}

	path += "*"
	c.GET(path, func(c Context) {
		fs.ServeHTTP(c.Response(), c.Request())
	})
}

// Run starts HTTP server
func (c *Chef) Run() {
	logger := c.logger.GetModuleLogger("chef")
	logger.Noticef("Running app on port %s", c.config.App.Port)
	logger.Fatal(http.ListenAndServe(c.config.App.Port, c.router))
}
