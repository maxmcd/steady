package daemon

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maxmcd/steady/daemon/api"
)

type server struct {
	daemon *Daemon
}

var _ api.ServerInterface = server{}

func (s server) GetApplication(c *gin.Context, name string) {
	s.daemon.applicationsLock.RLock()
	_, found := s.daemon.applications[name]
	s.daemon.applicationsLock.RUnlock()
	if !found {
		c.JSON(http.StatusNotFound, api.Error{Msg: "not found"})
		return
	}
	c.JSON(http.StatusOK, api.Application{Name: name})
}

func (s server) CreateApplication(c *gin.Context, name string) {
	var body api.CreateApplicationJSONBody
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, api.Error{Msg: err.Error()})
		return
	}

	if _, err := s.daemon.validateAndAddApplication(name, []byte(body.Script)); err != nil {
		c.JSON(http.StatusBadRequest, api.Error{Msg: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, api.Application{
		Name: name,
	})
}

func (s server) DeleteApplication(c *gin.Context, name string) {
	s.daemon.applicationsLock.RLock()
	_, found := s.daemon.applications[name]
	s.daemon.applicationsLock.RUnlock()
	if !found {
		c.JSON(http.StatusNotFound, api.Error{Msg: "not found"})
		return
	}

	s.daemon.applicationsLock.Lock()
	app := s.daemon.applications[name]
	delete(s.daemon.applications, name)
	s.daemon.applicationsLock.Unlock()

	if err := app.shutdown(); err != nil {
		c.JSON(http.StatusInternalServerError, api.Error{Msg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, api.Application{
		Name: name,
	})
}
