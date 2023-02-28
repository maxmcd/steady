// Package daemon provides the implementation of the Steady host daemon. The
// Daemon handles receiving application http requests, backing up sqlite
// databases, and managing the lifecycle of an application while it is on a
// host.
package daemon

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/benbjohnson/litestream"
	"github.com/benbjohnson/litestream/s3"
	"github.com/maxmcd/steady/daemon/daemonrpc"
	"github.com/maxmcd/steady/internal/boxpool"
	"github.com/maxmcd/steady/internal/httpx"
	"github.com/maxmcd/steady/internal/netx"
	_ "github.com/maxmcd/steady/internal/slogx"
	"github.com/maxmcd/steady/internal/steadyutil"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/samber/lo/parallel"
	"golang.org/x/exp/slog"
)

// Daemon is the steady daemon. It runs an http server, handles requests to
// create and delete applications, and handles database backups and migrations.
type Daemon struct {
	dataDirectory string
	client        *http.Client

	addr string

	s3Config *S3Config

	pool *boxpool.Pool

	applicationsLock sync.RWMutex
	applications     map[string]*application

	listener net.Listener
	wait     func() error
}

type DaemonOption func(*Daemon)

func NewDaemon(dataDirectory string, addr string, opts ...DaemonOption) *Daemon {
	boxpoolDir := filepath.Join(dataDirectory, "boxpool")
	if err := os.Mkdir(boxpoolDir, 0777); err != nil {
		panic(err)
	}
	pool, err := boxpool.New(context.Background(), "runner", boxpoolDir)
	if err != nil {
		panic(err)
	}

	d := &Daemon{
		dataDirectory: dataDirectory,
		addr:          addr,
		pool:          pool,
		applications:  map[string]*application{},
		client: &http.Client{
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: 1 * time.Second,
				}).Dial,
				ResponseHeaderTimeout: 10 * time.Second,
			},
		},
	}

	for _, opt := range opts {
		opt(d)
	}
	return d
}

type S3Config struct {
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	Path            string
	Endpoint        string
	SkipVerify      bool
	ForcePathStyle  bool
}

func DaemonOptionWithS3(cfg S3Config) DaemonOption { return func(d *Daemon) { d.s3Config = &cfg } }

// Wait for the server to exit, returning any errors.
func (d *Daemon) Wait() error {
	defer slog.Info("daemon stopped")
	if d.wait == nil {
		panic("Cannot call Addr when the server has not been started")
	}
	err := d.wait()
	d.applicationsLock.Lock()
	defer d.applicationsLock.Unlock()

	errs := parallel.Map(lo.Entries(d.applications),
		func(entry lo.Entry[string, *application], _ int) error {
			if err := entry.Value.shutdown(); err != nil {
				return fmt.Errorf("shutting down application %q: %w", entry.Key, err)
			}
			return nil
		},
	)
	for _, err := range errs {
		if err != nil {
			slog.Error(err.Error(), err)
		}
	}
	d.pool.Shutdown()
	return err
}

// ServerAddr returns the address of the running server. Will panic if the
// server hasn't been started yet.
func (d *Daemon) ServerAddr() string {
	if d.wait == nil {
		panic(fmt.Errorf("server has not started"))
	}
	return d.listener.Addr().String()
}

func (d *Daemon) s3Client() (*awss3.S3, error) {
	if d.s3Config == nil {
		panic("no s3 config")
	}
	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			d.s3Config.AccessKeyID,
			d.s3Config.SecretAccessKey,
			""),
		Endpoint:         aws.String(d.s3Config.Endpoint),
		Region:           aws.String("us-west-2"), // TODO
		DisableSSL:       aws.Bool(true),          // TODO
		S3ForcePathStyle: aws.Bool(d.s3Config.ForcePathStyle),
	}
	newSession, err := session.NewSession(s3Config)
	if err != nil {
		return nil, err
	}

	return awss3.New(newSession), err
}

func (d *Daemon) newReplica(db *litestream.DB, name string) *litestream.Replica {
	client := s3.NewReplicaClient()
	client.AccessKeyID = d.s3Config.AccessKeyID
	client.SecretAccessKey = d.s3Config.SecretAccessKey
	client.Bucket = d.s3Config.Bucket
	client.Path = filepath.Join(d.s3Config.Path, name)
	client.Endpoint = d.s3Config.Endpoint
	client.SkipVerify = d.s3Config.SkipVerify
	client.ForcePathStyle = d.s3Config.ForcePathStyle

	return litestream.NewReplica(db, name, client)
}

func (d *Daemon) createDB(name string) func(path string) (_ *litestream.DB, err error) {
	return func(path string) (_ *litestream.DB, err error) {
		if d.s3Config == nil {
			slog.Warn("skipping creating a new db because there is no s3 config")
			return nil, nil
		}
		db := litestream.NewDB(path)
		r := d.newReplica(db, filepath.Join(name, filepath.Base(path)))
		db.Replicas = append(db.Replicas, r)
		return db, nil
	}
}

func (d *Daemon) applicationHandler(name string, rw http.ResponseWriter, r *http.Request) {
	d.applicationsLock.RLock()
	app, found := d.applications[name]
	d.applicationsLock.RUnlock()
	if !found {
		http.Error(rw, "not found", http.StatusNotFound)
		return
	}

	// Need to run here to get app.box below
	if err := app.newRequest(r.Context()); err != nil {
		http.Error(rw, errors.Wrap(err, "error starting process").Error(), http.StatusInternalServerError)
		return
	}

	originalURL := r.URL
	// Route to correct port
	appURL := *r.URL
	appURL.Host = fmt.Sprintf("%s:%d", app.box.IPAddress(), app.port)
	appURL.Scheme = "http"
	r.URL = &appURL

	// Request.RequestURI can't be set in client requests.
	r.RequestURI = ""
	uri := r.RequestURI

	// TODO: Set to expected path for application to see
	// req.Header.Set("Host", "TODO")
	// TODO: x-forwarded-for

	defer func() {
		// Restore the uri for use with the echo logger middleware
		r.RequestURI = uri
		r.URL = originalURL
	}()

	defer app.endOfRequest()

	{
		resp, err := d.client.Do(r)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = r.Body.Close()

		for k, v := range resp.Header {
			rw.Header()[k] = v
		}
		rw.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(rw, resp.Body)
		_ = resp.Body.Close()
	}
}

// StopAllApplications will loop through all applications and shut them down.
// Useful for tests where we would otherwise wait for applications to shut down
// normally.
func (d *Daemon) StopAllApplications() {
	d.applicationsLock.Lock()
	for _, app := range d.applications {
		app.stopProcess(true)
	}
	d.applicationsLock.Unlock()
}

func (d *Daemon) downloadDatabasesIfFound(ctx context.Context, dir, name string) (err error) {
	dbs, err := d.findDatabasesForApplication(name)
	if err != nil {
		return errors.Wrap(err, "finding databases")
	}

	for _, db := range dbs {
		dbPath := filepath.Join(dir, db)

		// TODO: ensure we validate applications in an isolated environment
		// instead of just cleaning up after them
		//
		// TODO: also we just expect applications to run with an empty db? hmmm,
		// think about full lifecycle?
		_ = os.Remove(dbPath)

		replica := d.newReplica(nil, filepath.Join(name, db))
		generation, err := litestream.FindLatestGeneration(ctx, replica.Client())
		if err != nil {
			return errors.Wrap(err, "finding latest generation")
		}
		targetIndex, err := litestream.FindMaxIndexByGeneration(ctx, replica.Client(), generation)
		if err != nil {
			return errors.Wrap(err, "finding max index")
		}

		snapshotIndex, err := litestream.FindSnapshotForIndex(ctx, replica.Client(), generation, targetIndex)
		if err != nil {
			return errors.Wrap(err, "finding snapshot for index")
		}
		if err := litestream.Restore(ctx,
			replica.Client(),
			dbPath, generation,
			snapshotIndex, targetIndex,
			litestream.NewRestoreOptions()); err != nil {
			return errors.Wrap(err, "restoring")
		}
	}
	return nil
}

func (d *Daemon) findDatabasesForApplication(name string) (_ []string, err error) {
	s3Client, err := d.s3Client()
	if err != nil {
		return nil, err
	}
	output, err := s3Client.ListObjects(&awss3.ListObjectsInput{
		Bucket:    aws.String(d.s3Config.Bucket),
		Prefix:    aws.String(name + "/"),
		Delimiter: aws.String("/"),
	})
	if err != nil {
		return nil, err
	}

	var dbs []string
	for _, dir := range output.CommonPrefixes {
		dbName := strings.TrimSuffix(
			strings.TrimPrefix(
				*dir.Prefix,
				name+"/"),
			"/",
		)
		dbs = append(dbs, dbName)
	}
	return dbs, nil
}

func (d *Daemon) applicationDataDirectory() (loc string, err error) {
	return os.MkdirTemp(d.dataDirectory, "")
}

func (d *Daemon) updateApplication(name string, script []byte) error {
	d.applicationsLock.RLock()
	app, found := d.applications[name]
	d.applicationsLock.RUnlock()
	if !found {
		return fmt.Errorf("an application with this name is not present on this host")
	}
	if err := d.validateApplication(script); err != nil {
		return err
	}
	return app.updateApplication(script)
}

func (d *Daemon) validateApplication(script []byte) error {
	tmpDir, err := d.applicationDataDirectory()
	if err != nil {
		return err
	}
	fileName := filepath.Join(tmpDir, "index.ts")
	if err := os.WriteFile(fileName, script, 0600); err != nil {
		return fmt.Errorf("creating file %q: %w", fileName, err)
	}
	port, err := netx.GetFreePort()
	if err != nil {
		return err
	}
	box, err := bunRun(d.pool, tmpDir, port, nil)
	if err != nil {
		return err
	}
	_, _ = box.Stop()
	return os.RemoveAll(tmpDir)
}

func (d *Daemon) validateAndAddApplication(ctx context.Context, name string, script []byte) (*application, error) {
	// If two requests are validated simultaneously we won't catch them here,
	// but we'll error when adding them to the map later
	d.applicationsLock.RLock()
	_, found := d.applications[name]
	d.applicationsLock.RUnlock()
	if found {
		return nil, fmt.Errorf("an application with this name is already present on this host")
	}

	if err := d.validateApplication(script); err != nil {
		return nil, err
	}

	tmpDir, err := d.applicationDataDirectory()
	if err != nil {
		return nil, err
	}
	fileName := filepath.Join(tmpDir, "index.ts")
	if err := os.WriteFile(fileName, script, 0600); err != nil {
		return nil, fmt.Errorf("creating file %q: %w", fileName, err)
	}
	app := d.newApplication(name, tmpDir, 3000)
	app.waitForDB()
	app.start()
	d.applicationsLock.Lock()
	if _, found = d.applications[name]; found {
		d.applicationsLock.Unlock()
		return nil, fmt.Errorf("an application with this name is already present on this host")
	}
	d.applications[name] = app
	d.applicationsLock.Unlock()

	var dbs []string
	if d.s3Config != nil {
		if err := d.downloadDatabasesIfFound(ctx, tmpDir, name); err != nil {
			return nil, errors.Wrap(err, "downloading database")
		}
	}
	app.DBs = dbs
	if err := app.dbDownloaded(); err != nil {
		// TODO: recover!
		panic(err)
	}

	return app, nil
}

// Start starts the server.
func (d *Daemon) Start(ctx context.Context) error {
	if d.wait != nil {
		panic("Daemon has already started")
	}
	var err error
	d.listener, err = net.Listen("tcp", d.addr)
	if err != nil {
		return err
	}
	slog.Info("daemon listening", "addr", d.listener.Addr().String())

	daemonServer := daemonrpc.NewDaemonServer(server{daemon: d})

	d.wait = httpx.ServeContext(ctx, d.listener,
		&http.Server{
			Handler: steadyutil.Logger("dm", os.Stdout,
				http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
					name := steadyutil.ExtractAppName(r)
					if name == "" {
						http.Error(rw, "invalid host header", http.StatusBadRequest)
						return
					}
					if name == "steady" {
						daemonServer.ServeHTTP(rw, r)
					} else {
						d.applicationHandler(name, rw, r)
					}
				}),
			),
			ReadHeaderTimeout: time.Second * 15,
		},
	)
	return nil
}
