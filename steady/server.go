package steady

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/maxmcd/steady/daemon"
	db "github.com/maxmcd/steady/db"
	_ "github.com/maxmcd/steady/internal/slogx"
	"github.com/maxmcd/steady/steady/steadyrpc"
	"github.com/twitchtv/twirp"
)

type Server struct {
	db       *db.Queries
	dbClient *sqlx.DB

	daemonClient *daemon.Client

	privateLoadBalancerURL string
	publicLoadBalancerURL  string
	parsedPublicLB         *url.URL

	emailSink func(email string)
}

type ServerOptions struct {
	PrivateLoadBalancerURL string
	PublicLoadBalancerURL  string
	DaemonClient           *daemon.Client
}

func NewServer(options ServerOptions, opts ...Option) http.Handler {
	s := &Server{
		daemonClient:           options.DaemonClient,
		publicLoadBalancerURL:  options.PublicLoadBalancerURL,
		privateLoadBalancerURL: options.PrivateLoadBalancerURL,
	}
	var err error
	s.parsedPublicLB, err = url.Parse(s.publicLoadBalancerURL)
	if err != nil {
		panic(err)
	}
	if options.PublicLoadBalancerURL != "" &&
		s.parsedPublicLB.Scheme != "http" &&
		s.parsedPublicLB.Scheme != "https" {
		panic(fmt.Sprintf("Public load balancer url must be a full URL with schema: %q", options.PublicLoadBalancerURL))
	}

	for _, opt := range opts {
		opt(s)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token := r.Header.Get("X-Steady-Token"); token != "" {
			ctx, _ := twirp.WithHTTPRequestHeaders(r.Context(), http.Header{"X-Steady-Token": {token}})
			r = r.WithContext(ctx)
		}
		steadyrpc.NewSteadyServer(s).ServeHTTP(w, r)
	})
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
