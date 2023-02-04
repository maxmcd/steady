package steady

import (
	"context"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/maxmcd/steady/daemon"
	daemonrpc "github.com/maxmcd/steady/daemon/daemonrpc"
	db "github.com/maxmcd/steady/db"
	"github.com/maxmcd/steady/steady/steadyrpc"
)

type Server struct {
	db       *db.Queries
	dbClient *sqlx.DB

	daemonClient daemon.Client

	privateLoadBalancerHost string
	publicLoadBalancerURL   string
}

type ServerOptions struct {
	PrivateLoadBalancerURL string
	PublicLoadBalancerURL  string
	DaemonClient           daemon.Client
}

func NewServer(options ServerOptions, opts ...Option) *Server {
	s := &Server{
		daemonClient:            options.DaemonClient,
		publicLoadBalancerURL:   options.PublicLoadBalancerURL,
		privateLoadBalancerHost: options.PrivateLoadBalancerURL,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

var _ steadyrpc.Steady = new(Server)

func (s *Server) CreateService(ctx context.Context, req *steadyrpc.CreateServiceRequest) (
	_ *steadyrpc.CreateServiceResponse, err error) {
	service, err := s.db.CreateService(ctx, db.CreateServiceParams{
		Name:   req.Name,
		UserID: 1,
	})
	if err != nil {
		return nil, err
	}
	return &steadyrpc.CreateServiceResponse{
		Service: &steadyrpc.Service{
			Name:   service.Name,
			Id:     service.ID,
			UserId: service.UserID,
		},
	}, nil
}

type Option func(*Server)

func OptionWithSqlite(path string) Option {
	return func(s *Server) {
		dbClient, err := sqlx.Open("sqlite3", path)
		if err != nil {
			panic(err)
		}
		ctx := context.Background()
		tx, err := dbClient.BeginTxx(ctx, nil)
		if err != nil {
			panic(err)
		}
		{
			var userVersion int
			if err := tx.Get(&userVersion, "PRAGMA user_version"); err != nil {
				panic(err)
			}
			if userVersion != 1 {
				if _, err := tx.Exec(db.Migrations); err != nil {
					panic(err)
				}
			}
			if _, err := tx.Exec("PRAGMA user_version = 1"); err != nil {
				panic(err)
			}
		}
		if err := tx.Commit(); err != nil {
			panic(err)
		}
		s.db, err = db.Prepare(ctx, dbClient)
		if err != nil {
			panic(err)
		}
		s.dbClient = dbClient
	}
}

func OptionWithPostgres(connectionString string) Option {
	return func(s *Server) {
		dbClient, err := sqlx.Open("postgres", connectionString)
		if err != nil {
			panic(err)
		}
		// TODO: migration version
		if _, err := dbClient.Exec(db.Migrations); err != nil {
			panic(err)
		}
		s.dbClient = dbClient
	}
}

func (s *Server) dbTX(ctx context.Context) (db.Querier, error) {
	tx, err := s.dbClient.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return s.db.WithTx(tx), nil
}

func (s *Server) CreateServiceVersion(ctx context.Context, req *steadyrpc.CreateServiceVersionRequest) (
	_ *steadyrpc.CreateServiceVersionResponse, err error) {
	dbtx, err := s.dbTX(ctx)
	if err != nil {
		return nil, err
	}

	if _, err := dbtx.GetService(ctx, db.GetServiceParams{
		UserID: 1,
		ID:     req.ServiceId,
	}); err != nil {
		return nil, err
	}

	serviceVersion, err := dbtx.CreateServiceVersion(ctx, db.CreateServiceVersionParams{
		ServiceID: req.ServiceId,
		Version:   req.Version,
		Source:    req.Source,
	})
	if err != nil {
		return nil, err
	}

	return &steadyrpc.CreateServiceVersionResponse{
		ServiceVersion: &steadyrpc.ServiceVersion{
			Id:        serviceVersion.ID,
			ServiceId: serviceVersion.ServiceID,
			Version:   serviceVersion.Version,
			Source:    serviceVersion.Source,
		},
	}, nil
}

func (s *Server) DeployApplication(ctx context.Context, req *steadyrpc.DeployApplicationRequeast) (
	_ *steadyrpc.DeployApplicationResponse, err error) {
	serviceVersion, err := s.db.GetServiceVersion(ctx, req.ServiceVersionId)
	if err != nil {
		return nil, err
	}
	app, err := s.daemonClient.CreateApplication(ctx, &daemonrpc.CreateApplicationRequest{
		Name:   req.Name,
		Script: serviceVersion.Source,
	})
	if err != nil {
		return nil, err
	}

	// TODO: deploy application to host, confirm that it works
	return &steadyrpc.DeployApplicationResponse{
		Application: &steadyrpc.Application{Name: app.Name},
		Url:         s.publicLoadBalancerURL,
	}, nil
}

func (s *Server) DeploySource(ctx context.Context, req *steadyrpc.DeploySourceRequest) (
	_ *steadyrpc.DeploySourceResponse, err error) {
	app, err := s.daemonClient.CreateApplication(ctx, &daemonrpc.CreateApplicationRequest{
		Name:   "faketemporaryname",
		Script: req.Source,
	})
	if err != nil {
		return nil, err
	}
	_ = app
	// TODO: deploy application to host, confirm that it works
	return &steadyrpc.DeploySourceResponse{
		Url: app.Name,
	}, nil
}
