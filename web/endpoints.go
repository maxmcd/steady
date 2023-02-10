package web

import (
	"io"
	"mime"
	"path/filepath"

	"github.com/maxmcd/steady/steady/steadyrpc"
	"github.com/maxmcd/steady/web/mux"
)

func (s *Server) jsEditorAssetEndpoints(c *mux.Context) error {
	path := c.Params.ByName("path")
	c.Writer.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(path)))
	f, err := monacoSource.Open(
		filepath.Join(
			"node_modules/monaco-editor/min/",
			c.Params.ByName("path")))
	if err != nil {
		return err
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
