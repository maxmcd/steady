package web

import (
	"io"
	"mime"
	"path/filepath"

	"github.com/maxmcd/steady/steady/steadyrpc"
	"github.com/maxmcd/steady/web/mux"
)

func (s *Server) jsEditorAssetEndpoionts(c *mux.Context) error {
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
	c.Session.Values = map[interface{}]interface{}{}
	c.Session.AddFlash("You have been successfully logged out")
	c.SaveSession()
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
	c.Session.Values["user_id"] = resp.User.Id
	c.Session.Values["email"] = resp.User.Email
	c.Session.Values["username"] = resp.User.Username
	c.SaveSession()

	c.Redirect("/")
	return nil

}

func (s *Server) loginEndpoint(c *mux.Context) error {
	val := c.Request.FormValue("username_or_email")
	err := s.login(c.Request.Context(), val)
	if err != nil {
		return s.renderTemplateError(c, "login.go.html", V{"login_error": err.Error()}, err)
	}
	c.Session.AddFlash("An email with a login link is on its way to your inbox.")
	c.SaveSession()
	c.Redirect("/")
	return nil
}

func (s *Server) signupEndpoint(c *mux.Context) error {
	err := s.signup(c.Request.Context(),
		c.Request.FormValue("username"),
		c.Request.FormValue("email"),
	)
	if err != nil {
		return s.renderTemplateError(c, "login.go.html", V{"signup_error": err.Error()}, err)
	}
	c.Session.AddFlash("An email with a login link is on its way to your inbox.")
	c.SaveSession()
	c.Redirect("/")
	return nil
}
