package chef

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"sync"

	"github.com/gochef/cache"
	"github.com/gochef/session"
)

type (
	// Context is the current HTTP request context
	Context interface {
		SetHandlers(h []Handler)
		GetHandlers() []Handler
		Response() http.ResponseWriter
		Request() *http.Request
		Write(body []byte)
		WriteString(body string)
		JSON(data interface{}) error
		Param(key string) string
		FormValue(key string) string
		FormFile(key string) (*multipart.FileHeader, error)
		QueryString() string
		QueryParam(key string) string
		QueryParams() url.Values
		Set(key string, data interface{})
		Remove(key string)
		Get(key string) interface{}
		GetAll() Data
		GetInt(key string) int
		GetString(key string) string
		Redirect(location string, code int)
		Next()
		IsTLS() bool
		IsWebSocket() bool
		IsAjaxRequest() bool
		reset(req *http.Request, res http.ResponseWriter, config Config)
		File(file string) error
		SetStatusCode(code int)
		SetHeader(header, value string)
		Host() string
		Session() *session.Session
	}

	context struct {
		request   *http.Request
		response  http.ResponseWriter
		data      Data
		path      string
		pnames    []string
		pvalues   []string
		query     url.Values
		params    map[string]string
		handlers  []Handler
		next      Handler
		nextIndex int
		lock      sync.Mutex

		session *session.Session
		cache   *cache.Cache
	}
)

// NewContext returns a context instance
func NewContext(req *http.Request, res http.ResponseWriter, maxParam *int) Context {
	return &context{
		pvalues:  make([]string, *maxParam),
		params:   make(map[string]string),
		request:  req,
		response: res,
		data:     make(Data),
	}
}

func (c *context) SetHandlers(h []Handler) {
	c.handlers = h
}

func (c *context) GetHandlers() []Handler {
	return c.handlers
}

func (c *context) Response() http.ResponseWriter {
	return c.response
}

func (c *context) Request() *http.Request {
	return c.request
}

func (c *context) Write(body []byte) {
	c.response.Write(body)
}

func (c *context) WriteString(body string) {
	c.Write([]byte(body))
}

func (c *context) JSON(data interface{}) error {
	d, err := json.Marshal(data)
	if err != nil {
		return err
	}
	c.SetHeader(HeaderContentType, MIMEApplicationJSON)
	c.Write(d)
	return nil
}

func (c *context) Param(key string) string {
	return c.params[key]
}

func (c *context) FormValue(key string) string {
	return c.request.FormValue(key)
}

func (c *context) FormFile(key string) (*multipart.FileHeader, error) {
	_, fh, err := c.request.FormFile(key)
	return fh, err
}

func (c *context) QueryString() string {
	return c.request.URL.RawQuery
}

func (c *context) QueryParam(key string) string {
	return c.QueryParams().Get(key)
}

func (c *context) QueryParams() url.Values {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	return c.query
}

func (c *context) Set(key string, data interface{}) {
	c.lock.Lock()
	if c.data == nil {
		c.data = make(Data)
	}

	c.data[key] = data
	c.lock.Unlock()
}

func (c *context) Remove(key string) {
	c.lock.Lock()
	delete(c.data, key)
	c.lock.Unlock()
}

func (c *context) Get(key string) interface{} {
	return c.data[key]
}

func (c *context) GetAll() Data {
	return c.data
}

func (c *context) GetInt(key string) int {
	data := c.Get(key)
	if data == nil {
		return 0
	}

	if d, ok := data.(int); ok {
		return d
	}

	return 0
}

func (c *context) GetString(key string) string {
	data := c.Get(key)
	if data == nil {
		return ""
	}

	if d, ok := data.(string); ok {
		return d
	}

	return ""
}

func (c *context) Redirect(location string, code int) {
	http.Redirect(c.response, c.request, location, code)
}

func (c *context) reset(req *http.Request, res http.ResponseWriter, config Config) {
	c.nextIndex = -1
	c.request = req
	c.response = res
	c.path = ""
	c.pnames = nil
	c.handlers = []Handler{
		NotFoundHandler,
	}

	if config.Session.Use {
		c.session = session.GetDriver(config.Session, req, res)
	}

	if config.Cache.Use {
		c.cache = cache.GetDriver(config.Cache)
	}
}

func (c *context) Next() {
	c.nextIndex++
	lenHandlers := len(c.handlers)

	if (lenHandlers > 0) && (c.nextIndex < lenHandlers) {
		c.handlers[c.nextIndex](c)
	}
}

func (c *context) IsTLS() bool {
	return c.request.TLS != nil
}

func (c *context) IsWebSocket() bool {
	return false
}

func (c *context) IsAjaxRequest() bool {
	return c.request.Header.Get(HeaderXRequestedWith) == MIMEApplicationAjax
}

func (c *context) File(file string) error {
	f, err := os.Open(file)
	if err != nil {
		NotFoundHandler(c)
		return err
	}
	defer f.Close()

	fi, _ := f.Stat()
	if !fi.IsDir() {
		http.ServeContent(c.response, c.request, fi.Name(), fi.ModTime(), f)
	}
	return nil
}

func (c *context) SetStatusCode(code int) {
	c.response.WriteHeader(code)
}

func (c *context) SetHeader(header, value string) {
	c.response.Header().Set(header, value)
}

func (c *context) Host() string {
	return c.request.Host
}

func (c *context) Session() *session.Session {
	return c.session
}
