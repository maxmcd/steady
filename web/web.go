package web

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/maxmcd/steady/steady/steadyrpc"
	"github.com/maxmcd/steady/web/mux"
	"github.com/pkg/errors"
	"github.com/twitchtv/twirp"
)

//go:generate bun install

//go:embed templates/*
var templates embed.FS

//go:embed node_modules/monaco-editor/min/vs/*
var monacoSource embed.FS

type Server struct {
	t            *template.Template
	router       *mux.Router
	steadyClient steadyrpc.Steady
}

func NewServer(steadyClient steadyrpc.Steady) (http.Handler, error) {
	t, err := template.ParseFS(templates, "templates/*")
	if err != nil {
		return nil, errors.Wrap(err, "error running ParseFS")
	}

	s := &Server{
		t:            t,
		steadyClient: steadyClient,
		router: mux.NewRouter(
			// TODO: must be secure to avoid spoofing
			sessions.NewCookieStore([]byte("TODO")),
		),
	}
	s.router.GET("/", func(c *mux.Context) error { return s.renderTemplate(c, "index.go.html", nil) })
	s.router.GET("/js/editor/*path", s.jsEditorAssetEndpoionts)
	s.router.GET("/login", func(c *mux.Context) error {
		if s.loggedIn(c) {
			c.Redirect("/")
			return nil
		}
		return s.renderTemplate(c, "login.go.html", nil)
	})
	s.router.GET("/logout", s.logoutEndpoint)
	s.router.GET("/login/token/:token", s.tokenLoginEndpoint)
	s.router.POST("/login", s.loginEndpoint)
	s.router.POST("/signup", s.signupEndpoint)

	return s.router, nil
}

type V map[string]interface{}

type pageData map[string]interface{}

func (pd pageData) LoggedIn() bool {
	if session, found := pd["session"]; found {
		_, found = session.(map[interface{}]interface{})["user_id"]
		return found
	}
	return false
}

func (s *Server) renderTemplateError(c *mux.Context, name string, data map[string]interface{}, err error) error {
	if er, match := err.(twirp.Error); match {
		c.Writer.WriteHeader(twirp.ServerHTTPStatusFromErrorCode(er.Code()))
	}
	return s.renderTemplate(c, name, data)
}

func (s *Server) loggedIn(c *mux.Context) bool {
	_, found := c.Session.Values["user_id"]
	return found
}

func (s *Server) renderTemplate(c *mux.Context, name string, data map[string]interface{}) error {
	if data == nil {
		data = map[string]interface{}{}
	}
	data["flashes"] = c.Session.Flashes()
	data["session"] = c.Session.Values
	fmt.Println(data)
	c.SaveSession()
	return s.t.Lookup(name).Execute(c.Writer, pageData(data))
}

func (s *Server) login(ctx context.Context, usernameOrEmail string) error {
	if usernameOrEmail == "" {
		return twirp.NewError(twirp.InvalidArgument, "username or email cannot be blank")
	}
	_, err := s.steadyClient.Login(ctx, &steadyrpc.LoginRequest{
		Username: usernameOrEmail,
		Email:    usernameOrEmail,
	})
	return err
}

func (s *Server) signup(ctx context.Context, username, email string) error {
	if username == "" {
		return twirp.NewError(twirp.InvalidArgument, "username cannot be blank")
	}
	if email == "" {
		return twirp.NewError(twirp.InvalidArgument, "email cannot be blank")
	}
	_, err := s.steadyClient.Signup(ctx, &steadyrpc.SignupRequest{
		Username: username,
		Email:    email,
	})
	return err
}
