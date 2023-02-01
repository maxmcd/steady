package steady

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	db "github.com/maxmcd/steady/db"
	"github.com/maxmcd/steady/steady/rpc"
)

type Service struct {
	db db.Querier
}

type Option func(*Service)

func OptionWithSqlite(path string) Option {
	return func(s *Service) {
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
		s.db = db.New(dbClient)
	}
}
func OptionWithPostgres(connectionString string) Option {
	return func(s *Service) {
		dbClient, err := sql.Open("postgres", connectionString)
		if err != nil {
			panic(err)
		}
		// TODO: migration version
		if _, err := dbClient.Exec(db.Migrations); err != nil {
			panic(err)
		}
		s.db = db.New(dbClient)
	}
}

var _ rpc.Steady = new(Service)

func (s *Service) CreateService(ctx context.Context, req *rpc.CreateServiceRequest) (
	_ *rpc.CreateServiceResponse, err error) {
	service, err := s.db.CreateService(ctx, db.CreateServiceParams{
		Name:   req.Name,
		UserID: 1,
	})
	if err != nil {
		return nil, err
	}
	return &rpc.CreateServiceResponse{
		Service: &rpc.Service{
			Name:   service.Name,
			Id:     service.ID,
			UserId: service.UserID,
		},
	}, nil
}

func (s *Service) CreateServiceVersion(ctx context.Context, req *rpc.CreateServiceVersionRequest) (
	_ *rpc.CreateServiceVersionResponse, err error) {
	return nil, err
}

func (s *Service) DeployApplication(ctx context.Context, req *rpc.DeployApplicationRequeast) (
	_ *rpc.DeployApplicationResponse, err error) {
	return nil, err
}
