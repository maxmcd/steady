package mux

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/julienschmidt/httprouter"
	"github.com/twitchtv/twirp"
)

type Router struct {
	router *httprouter.Router
	store  sessions.Store
}

func NewRouter(store sessions.Store) *Router {
	return &Router{
		router: httprouter.New(),
		store:  store,
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

func (r *Router) handlerWrapper(handler func(c *Context) error) func(
	http.ResponseWriter, *http.Request, httprouter.Params) {
	return func(w http.ResponseWriter, req *http.Request, p httprouter.Params) {
		c := &Context{
			Writer:  w,
			Request: req,
			Params:  p,
		}
		c.Session, _ = r.store.Get(req, "session")
		err := handler(c)
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

func (r *Router) GET(path string, handler func(c *Context) error) {
	r.router.Handle(http.MethodGet, path, r.handlerWrapper(handler))
}

func (r *Router) POST(path string, handler func(c *Context) error) {
	r.router.Handle(http.MethodPost, path, r.handlerWrapper(handler))
}

type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request
	Session *sessions.Session
	Params  httprouter.Params
}

func (c *Context) Redirect(path string) {
	http.Redirect(c.Writer, c.Request, path, http.StatusFound)
}
func (c *Context) SaveSession() {
	if err := c.Session.Save(c.Request, c.Writer); err != nil {
		// I believe this will just error in incorrect types
		panic(err)
	}
}
