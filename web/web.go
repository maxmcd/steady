package web

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"net"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/julienschmidt/httprouter"
	"github.com/maxmcd/steady/steady/steadyrpc"
	"github.com/pkg/errors"
	"github.com/twitchtv/twirp"
)

//go:embed templates/*
var templates embed.FS

type Server struct {
	t            *template.Template
	router       *httprouter.Router
	steadyClient steadyrpc.Steady
	store        sessions.Store
}

func NewServer(steadyClient steadyrpc.Steady) (*Server, error) {
	t, err := template.ParseFS(templates, "*/*")
	if err != nil {
		return nil, errors.Wrap(err, "error running ParseFS")
	}

	s := &Server{
		router:       httprouter.New(),
		t:            t,
		steadyClient: steadyClient,
		// TODO: must be secure to avoid spoofing
		store: sessions.NewCookieStore([]byte("TODO")),
	}
	s.get("/", func(c *Context) error {
		return c.renderTemplate("index.go.html", nil)
	})
	s.get("/login", func(c *Context) error {
		return c.renderTemplate("login.go.html", nil)
	})
	s.get("/login/token/:token", func(c *Context) error {
		resp, err := s.steadyClient.ValidateToken(c.req.Context(), &steadyrpc.ValidateTokenRequest{
			Token: c.params.ByName("token"),
		})
		if err != nil {
			return err
		}
		c.session.Values["user_id"] = resp.User.Id
		c.session.Values["email"] = resp.User.Email
		c.session.Values["username"] = resp.User.Username
		if err := c.saveSession(); err != nil {
			return err
		}

		c.redirect("/")
		return nil
	})
	s.post("/login", func(c *Context) error {
		val := c.req.FormValue("username_or_email")
		err := s.login(c.req.Context(), val)
		if err != nil {
			return c.renderTemplate("login.go.html", V{"login_error": err.Error()})
		}
		c.redirect("/")
		return nil
	})
	s.post("/signup", func(c *Context) error {
		val := c.req.FormValue("username_or_email")
		err := s.signup(c.req.Context(), val)
		if err != nil {
			return c.renderTemplate("login.go.html", V{"signup_error": err.Error()})
		}
		c.redirect("/")
		return nil
	})
	return s, nil
}

func (s *Server) Run(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	srv := &http.Server{
		Handler: s.router,
	}
	fmt.Printf("Listening on %s\n", listener.Addr())
	return srv.Serve(listener)
}

type V map[string]interface{}

type Context struct {
	writer  http.ResponseWriter
	req     *http.Request
	session *sessions.Session
	params  httprouter.Params

	s *Server
}

func (c *Context) renderTemplate(name string, data map[string]interface{}) error {
	if data == nil {
		data = map[string]interface{}{}
	}
	data["flashes"] = c.session.Flashes()
	data["session"] = c.session.Values
	fmt.Println(data)
	return c.s.t.Lookup(name).Execute(c.writer, data)
}
func (c *Context) redirect(path string) {
	http.Redirect(c.writer, c.req, path, http.StatusFound)
}
func (c *Context) saveSession() error {
	return c.session.Save(c.req, c.writer)
}

func (s *Server) handlerWrapper(handler func(c *Context) error) func(
	w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		c := &Context{
			writer: w,
			req:    r,
			params: p,
			s:      s,
		}
		c.session, _ = s.store.Get(r, "session")
		fmt.Println(c.session.Flashes(), c.session.Values)
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

func (s *Server) get(path string, handler func(c *Context) error) {
	s.router.Handle(http.MethodGet, path, s.handlerWrapper(handler))
}

func (s *Server) post(path string, handler func(c *Context) error) {
	s.router.Handle(http.MethodPost, path, s.handlerWrapper(handler))
}

var ErrBlankUsernameOrEmail = fmt.Errorf("username or email cannot be blank")

func (s *Server) login(ctx context.Context, usernameOrEmail string) error {
	if usernameOrEmail == "" {
		return ErrBlankUsernameOrEmail
	}
	_, err := s.steadyClient.Login(ctx, &steadyrpc.LoginRequest{
		Username: usernameOrEmail,
		Email:    usernameOrEmail,
	})
	return err
}

func (s *Server) signup(ctx context.Context, usernameOrEmail string) error {
	if usernameOrEmail == "" {
		return ErrBlankUsernameOrEmail
	}
	_, err := s.steadyClient.Signup(ctx, &steadyrpc.SignupRequest{
		Username: usernameOrEmail,
		Email:    usernameOrEmail,
	})
	return err
}
