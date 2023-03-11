package web

import (
	"io"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/maxmcd/steady/internal/mux"
	"github.com/maxmcd/steady/steady/steadyrpc"
	"github.com/twitchtv/twirp"
	"golang.org/x/exp/slog"
)

func (s *Server) assetsEndpoints(c *mux.Context) error {
	path := c.Params.ByName("path")
	slog.Info(path)
	if path == "/bun-types/types.d.ts" {
		return s.bunTypesEndpoint(c)
	}
	c.Writer.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(path)))
	f, err := distFiles.Open(
		filepath.Join("app/dist/", c.Params.ByName("path")))
	if err != nil {
		return twirp.NotFoundError("not found")
	}

	_, _ = io.Copy(c.Writer, f)
	return nil
}

func (s *Server) bunTypesEndpoint(c *mux.Context) error {
	c.Writer.Header().Set("Content-Type", mime.TypeByExtension(".ts"))
	c.Writer.Write(bunTypes)
	return nil
}

func (s *Server) logoutEndpoint(c *mux.Context) error {
	ctx, err := twirp.WithHTTPRequestHeaders(c.Request.Context(), http.Header{
		"X-Steady-Token": []string{c.Token},
	})
	if err != nil {
		return err
	}
	if _, err := s.steadyClient.Logout(ctx, nil); err != nil {
		return err
	}

	c.DeleteToken()
	c.AddFlash("You have been successfully logged out")
	c.SaveFlash()
	c.Redirect("/")
	return nil
}

func (s *Server) tokenLoginEndpoint(c *mux.Context) error {
	resp, err := s.steadyClient.ValidateToken(c.Request.Context(), &steadyrpc.ValidateTokenRequest{
		Token: c.Params.ByName("token"),
	})
	if err != nil {
		return err
	}
	c.SetToken(resp.UserSessionToken)
	c.Redirect("/")
	return nil
}

func (s *Server) loginEndpoint(c *mux.Context) error {
	val := c.Request.FormValue("username_or_email")
	err := s.login(c.Request.Context(), val)
	if err != nil {
		c.Data["username_or_email"] = val
		return NewTemplateError(err, "login.go.html", "login_error", twirp.InvalidArgument)
	}
	c.AddFlash("An email with a login link is on its way to your inbox.")
	c.SaveFlash()
	c.Redirect("/")
	return nil
}

func (s *Server) signupEndpoint(c *mux.Context) error {
	username := c.Request.FormValue("username")
	email := c.Request.FormValue("email")
	err := s.signup(c.Request.Context(), username, email)
	if err != nil {
		c.Data["username"] = username
		c.Data["email"] = email
		return NewTemplateError(err, "login.go.html", "signup_error", twirp.InvalidArgument)
	}
	c.AddFlash("An email with a login link is on its way to your inbox.")
	c.SaveFlash()
	c.Redirect("/")
	return nil
}

func (s *Server) runApplication(c *mux.Context) error {
	source := c.Request.FormValue("index.ts")
	resp, err := s.steadyClient.RunApplication(c.Request.Context(), &steadyrpc.RunApplicationRequest{
		Source: source,
	})
	if err != nil {
		return NewTemplateError(
			err, "index.go.html",
			"application_error", twirp.InvalidArgument)
	}

	c.Redirect("/app", resp.Application.Name)
	return nil
}

func (s *Server) updateApplication(c *mux.Context) error {
	source := c.Request.FormValue("index.ts")
	resp, err := s.steadyClient.RunApplication(c.Request.Context(), &steadyrpc.RunApplicationRequest{
		Source: source,
	})
	if err != nil {
		return NewTemplateError(
			err, "index.go.html",
			"application_error", twirp.InvalidArgument)
	}

	c.Redirect("/app", resp.Application.Name)
	return nil
}

func (s *Server) showApplication(c *mux.Context) error {
	appName := c.Params.ByName("name")

	resp, err := s.steadyClient.GetApplication(c.Request.Context(), &steadyrpc.GetApplicationRequest{Name: appName})
	if err != nil {
		return err
	}

	c.Data["app"] = resp.Application
	c.Data["app_url"] = resp.Url
	return s.renderTemplate(c, "application.go.html")
}

type TemplateError struct {
	err       error
	errorCode twirp.ErrorCode
	template  string
	errorName string
}

func NewTemplateError(err error, template string, name string, code twirp.ErrorCode) TemplateError {
	return TemplateError{
		errorCode: code,
		err:       err,
		template:  template,
		errorName: name,
	}
}

func (t TemplateError) Error() string {
	if er, match := t.err.(twirp.Error); match {
		return er.Msg()
	}
	return t.err.Error()
}
