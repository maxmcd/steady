package steady

import (
	"context"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/maxmcd/steady/daemon"
	db "github.com/maxmcd/steady/db"
	"github.com/maxmcd/steady/steady/steadyrpc"
)

type Server struct {
	db       *db.Queries
	dbClient *sqlx.DB

	daemonClient daemon.Client

	privateLoadBalancerHost string
	publicLoadBalancerURL   string

	emailSink func(email string)
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

func OptionWithEmailSink(e func(email string)) Option {
	return func(s *Server) {
		s.emailSink = e
	}
}

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
