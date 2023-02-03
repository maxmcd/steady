package daemon

import (
	"context"

	"github.com/maxmcd/steady/daemon/rpc"
	"github.com/twitchtv/twirp"
)

type server struct {
	daemon *Daemon
}

var _ rpc.Daemon = new(server)

func (s server) CreateApplication(ctx context.Context, req *rpc.CreateApplicationRequest) (
	_ *rpc.Application, err error) {
	if _, err := s.daemon.validateAndAddApplication(ctx, req.Name, []byte(req.Script)); err != nil {
		return nil, twirp.NewError(twirp.InvalidArgument, err.Error())
	}
	return &rpc.Application{Name: req.Name}, nil
}

func (s server) DeleteApplication(ctx context.Context, req *rpc.DeleteApplicationRequest) (
	_ *rpc.Application, err error) {
	name := req.Name
	s.daemon.applicationsLock.RLock()
	_, found := s.daemon.applications[name]
	s.daemon.applicationsLock.RUnlock()
	if !found {
		return nil, twirp.NewError(twirp.NotFound, "not found")
	}

	s.daemon.applicationsLock.Lock()
	app := s.daemon.applications[name]
	delete(s.daemon.applications, name)
	s.daemon.applicationsLock.Unlock()

	if err := app.shutdown(); err != nil {
		return nil, twirp.NewError(twirp.Internal, err.Error())
	}

	return &rpc.Application{
		Name:         name,
		RequestCount: int64(app.requestCount),
		StartCount:   int64(app.startCount),
	}, nil
}

func (s server) GetApplication(ctx context.Context, req *rpc.GetApplicationRequest) (
	_ *rpc.Application, err error) {
	s.daemon.applicationsLock.RLock()
	app, found := s.daemon.applications[req.Name]
	s.daemon.applicationsLock.RUnlock()
	if !found {
		return nil, twirp.NewError(twirp.NotFound, "not found")
	}
	return &rpc.Application{
		Name:         req.Name,
		RequestCount: int64(app.requestCount),
		StartCount:   int64(app.startCount),
	}, nil
}
