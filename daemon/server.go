package daemon

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maxmcd/steady/daemon/api"
)

type server struct {
	daemon *Daemon
}

type HTTPError struct {
	code int
	err  error
}

func (s server) GetApplication(c *gin.Context, name string) {
	s.daemon.applicationsLock.RLock()
	app, found := s.daemon.applications[name]
	s.daemon.applicationsLock.RUnlock()
	if !found {
		c.JSON(http.StatusNotFound, api.Error{Msg: "not found"})
		return
	}
	c.JSON(http.StatusOK, api.Application{Name: name, Port: app.port})
}

func (s server) CreateApplication(c *gin.Context, name string) {
	var body api.CreateApplicationJSONBody
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, api.Error{Msg: err.Error()})
		return
	}
	app, err := s.daemon.validateAndAddApplication(name, []byte(body.Script))
	if err != nil {
		c.JSON(http.StatusBadRequest, api.Error{Msg: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, api.Application{
		Name: name,
		Port: app.port,
	})
}
