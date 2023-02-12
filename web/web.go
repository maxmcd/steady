package web

import (
	"context"
	"embed"
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/maxmcd/steady/internal/steadyutil"
	"github.com/maxmcd/steady/steady/steadyrpc"
	"github.com/maxmcd/steady/web/mux"
	"github.com/pkg/errors"
	"github.com/twitchtv/twirp"
)

//go:embed templates/*
var templates embed.FS

//go:embed dist/*
var distFiles embed.FS

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

	if t, err = t.ParseFS(distFiles, "dist/_assets.go.html"); err != nil {
		return nil, errors.Wrap(err, "error running ParseFS 2")
	}

	s := &Server{
		t:            t,
		steadyClient: steadyClient,
		router:       mux.NewRouter(),
	}
	s.router.GET("/", func(c *mux.Context) error {
		return s.renderTemplate(c, "index.go.html")
	})
	s.router.GET("/assets/*path", s.assetsEndpoints)
	s.router.GET("/login", func(c *mux.Context) error {
		if s.loggedIn(c) {
			c.Redirect("/")
			return nil
		}
		return s.renderTemplate(c, "login.go.html")
	})
	s.router.GET("/logout", s.logoutEndpoint)
	s.router.GET("/login/token/:token", s.tokenLoginEndpoint)
	s.router.POST("/login", s.loginEndpoint)
	s.router.POST("/signup", s.signupEndpoint)

	s.router.POST("/application", s.runApplication)
	s.router.GET("/application/:name", s.showApplication)

	return s.router, nil
}

func WebAndSteadyServer(steadyHandler, webHandler http.Handler) http.Handler {
	return steadyutil.Logger(os.Stdout,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/twirp") {
				steadyHandler.ServeHTTP(w, r)
			} else {
				webHandler.ServeHTTP(w, r)
			}
		}),
	)
}

type V map[string]interface{}

type pageData map[string]interface{}

func (pd pageData) LoggedIn() bool {
	if user, found := pd["user"]; found {
		_, match := user.(*steadyrpc.User)
		return match
	}
	return false
}

func tidyErrorMessage(err error) string {
	if er, match := err.(twirp.Error); match {
		return er.Msg()
	}
	return err.Error()
}

func (s *Server) renderTemplateError(c *mux.Context, name string, err error) error {
	if er, match := err.(twirp.Error); match {
		c.Writer.WriteHeader(twirp.ServerHTTPStatusFromErrorCode(er.Code()))
	}
	return s.renderTemplate(c, name)
}

func (s *Server) getUser(c *mux.Context) (*steadyrpc.User, error) {
	if !s.loggedIn(c) {
		return nil, twirp.NewError(twirp.Unauthenticated, "unauthenticated")
	}
	if user, found := c.Data["user"]; found {
		return user.(*steadyrpc.User), nil
	}
	ctx, err := twirp.WithHTTPRequestHeaders(c.Request.Context(), http.Header{
		"X-Steady-Token": []string{c.Token},
	})
	if err != nil {
		return nil, err
	}
	resp, err := s.steadyClient.GetUser(ctx, nil)
	if err != nil {
		return nil, err
	}
	c.Data["user"] = resp.User
	return resp.User, nil
}

func (s *Server) loggedIn(c *mux.Context) bool {
	return c.Token != ""
}

func (s *Server) renderTemplate(c *mux.Context, name string) error {
	if s.loggedIn(c) {
		if _, err := s.getUser(c); err != nil {
			return err
		}
	}
	data := c.Data
	data["flashes"] = c.GetFlashes()
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
	if err != nil {
		if strings.Contains(err.Error(), "not_found") {
			return errors.New("A user with this username or email could not be found")
		}
	}
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
