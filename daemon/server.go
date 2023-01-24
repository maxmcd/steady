package daemon

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/maxmcd/steady/daemon/api"
)

//go:generate bash -c "oapi-codegen --package=api --generate=types,client,server,spec ./openapi3.yaml > api/api.gen.go"

type server struct {
	daemon *Daemon
}

var _ api.ServerInterface = server{}

func (s server) GetApplication(c echo.Context, name string) error {
	s.daemon.applicationsLock.RLock()
	app, found := s.daemon.applications[name]
	s.daemon.applicationsLock.RUnlock()
	if !found {
		return echo.NewHTTPError(http.StatusNotFound, api.Error{Msg: "not found"})
	}
	return c.JSON(http.StatusOK, api.Application{
		Name:         name,
		RequestCount: app.requestCount,
		StartCount:   app.startCount,
	})
}

func (s server) CreateApplication(c echo.Context, name string) error {
	var body api.CreateApplicationJSONBody
	if err := c.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, api.Error{Msg: err.Error()})
	}

	if _, err := s.daemon.validateAndAddApplication(name, []byte(body.Script)); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, api.Error{Msg: err.Error()})
	}

	return c.JSON(http.StatusCreated, api.Application{
		Name: name,
	})
}

func (s server) DeleteApplication(c echo.Context, name string) error {
	s.daemon.applicationsLock.RLock()
	_, found := s.daemon.applications[name]
	s.daemon.applicationsLock.RUnlock()
	if !found {
		return echo.NewHTTPError(http.StatusNotFound, api.Error{Msg: "not found"})
	}

	s.daemon.applicationsLock.Lock()
	app := s.daemon.applications[name]
	delete(s.daemon.applications, name)
	s.daemon.applicationsLock.Unlock()

	if err := app.shutdown(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, api.Error{Msg: err.Error()})
	}

	return c.JSON(http.StatusOK, api.Application{
		Name: name,
	})
}
