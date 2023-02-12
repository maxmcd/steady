package web

import (
	"fmt"
	"io"
	"mime"
	"path/filepath"

	"github.com/maxmcd/steady/steady/steadyrpc"
	"github.com/maxmcd/steady/web/mux"
	"github.com/twitchtv/twirp"
)

func (s *Server) assetsEndpoints(c *mux.Context) error {
	path := c.Params.ByName("path")
	c.Writer.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(path)))
	fmt.Println(filepath.Join(
		"dist/",
		c.Params.ByName("path")))
	f, err := distFiles.Open(
		filepath.Join(
			"dist/",
			c.Params.ByName("path")))
	if err != nil {
		return twirp.NotFoundError("not found")
	}

	_, _ = io.Copy(c.Writer, f)
	return nil
}

func (s *Server) logoutEndpoint(c *mux.Context) error {
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
		return NewTemplateError("login.go.html", "login_error", twirp.InvalidArgument, err.Error())
		c.Data["login_error"] = tidyErrorMessage(err)
		return s.renderTemplateError(c, "login.go.html", err)
	}
	c.AddFlash("An email with a login link is on its way to your inbox.")
	c.SaveFlash()
	c.Redirect("/")
	return nil
}

func (s *Server) signupEndpoint(c *mux.Context) error {
	err := s.signup(c.Request.Context(),
		c.Request.FormValue("username"),
		c.Request.FormValue("email"),
	)
	if err != nil {
		c.Data["signup_error"] = tidyErrorMessage(err)
		return s.renderTemplateError(c, "login.go.html", err)
	}
	c.AddFlash("An email with a login link is on its way to your inbox.")
	c.SaveFlash()
	c.Redirect("/")
	return nil
}

func (s *Server) runApplication(c *mux.Context) error {
	source := c.Request.FormValue("index.ts")
	if source == "" {
		return twirp.NewError(twirp.InvalidArgument, "Application source cannot be empty")
	}
	resp, err := s.steadyClient.RunApplication(c.Request.Context(), &steadyrpc.RunApplicationRequest{
		Source: &source,
	})
	if err != nil {
		return err
	}

	c.Redirect("/application", resp.Application.Name)
	return nil
}

func (s *Server) showApplication(c *mux.Context) error {
	appName := c.Params.ByName("name")

	resp, err := s.steadyClient.GetApplication(c.Request.Context(), &steadyrpc.GetApplicationRequest{Name: appName})
	if err != nil {
		return err
	}

	c.Data["app"] = resp.Application
	return s.renderTemplate(c, "application.go.html")
}

type TemplateError struct {
	errorCode twirp.ErrorCode
	msg       string
	template  string
	errorName string
}

func NewTemplateError(template string, name string, code twirp.ErrorCode, msg string) TemplateError {
	return TemplateError{
		errorCode: code,
		msg:       msg,
		template:  template,
		errorName: name,
	}
}

func (t TemplateError) Error() string {
	return t.msg
}
