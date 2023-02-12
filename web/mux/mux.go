package mux

import (
	"net/http"
	"path"

	"github.com/julienschmidt/httprouter"
	"github.com/twitchtv/twirp"
)

type Router struct {
	router       *httprouter.Router
	ErrorHandler func(c *Context, err error)
}

func NewRouter() *Router {
	return &Router{
		router: httprouter.New(),
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

func (r *Router) handlerWrapper(handler func(c *Context) error) func(
	http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
		c := &Context{
			// TODO: wrap writer so that we can automatically save flashes
			Writer:  w,
			Request: req,
			Params:  p,
			Token:   getToken(w, req),
			Data:    map[string]interface{}{},
		}
		err := handler(c)
		if r.ErrorHandler != nil {
			r.ErrorHandler(c, err)
		} else {
			if err != nil {
				code, msg := http.StatusInternalServerError, err.Error()
				if er, match := err.(twirp.Error); match {
					code = twirp.ServerHTTPStatusFromErrorCode(er.Code())
					msg = er.Msg()
				}
				http.Error(w, msg, code)
			}
		}
	}
}

func (r *Router) GET(path string, handler func(c *Context) error) {
	r.router.Handle(http.MethodGet, path, r.handlerWrapper(handler))
}

func (r *Router) POST(path string, handler func(c *Context) error) {
	r.router.Handle(http.MethodPost, path, r.handlerWrapper(handler))
}

type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request
	Params  httprouter.Params
	flashes []string
	Token   string
	Data    map[string]interface{}
}

func (c *Context) Redirect(paths ...string) {
	http.Redirect(c.Writer, c.Request, path.Join(paths...), http.StatusFound)
}

func (c *Context) AddFlash(msg string) {
	c.flashes = append(c.flashes, msg)
}
func (c *Context) SaveFlash() {
	setFlash(c.Writer, c.flashes)
}

func (c *Context) GetFlashes() []string {
	return getFlash(c.Writer, c.Request)
}

func (c *Context) SetToken(token string) {
	setToken(c.Writer, token)
}

func (c *Context) DeleteToken() {
	deleteToken(c.Writer, c.Request)
}
