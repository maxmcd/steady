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
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/maxmcd/steady/daemon/api"
	"github.com/maxmcd/steady/internal/netx"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type Daemon struct {
	dataDirectory string
	addr          string
	client        *http.Client

	s3Config *S3Config

	applicationsLock sync.RWMutex
	applications     map[string]*Application

	listener     net.Listener
	listenerWait *sync.WaitGroup
	eg           *errgroup.Group
}

type DaemonOption func(*Daemon)

func NewDaemon(dataDirectory string, addr string, opts ...DaemonOption) *Daemon {
	d := &Daemon{
		dataDirectory: dataDirectory,
		addr:          addr,
		applications:  map[string]*Application{},
		listenerWait:  &sync.WaitGroup{},
		client: &http.Client{
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: 1 * time.Second,
				}).Dial,
				ResponseHeaderTimeout: 10 * time.Second,
			},
		},
	}
	d.listenerWait.Add(1)

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

func (d *Daemon) Wait() error { return d.eg.Wait() }

func (d *Daemon) ServerAddr() string {
	if d.eg == nil {
		panic(fmt.Errorf("server has not started"))
	}
	d.listenerWait.Wait()
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
			fmt.Println("WARN: skipping creating a new db because there is no s3 config")
			return nil, nil
		}
		db := litestream.NewDB(path)
		r := d.newReplica(db, filepath.Join(name, filepath.Base(path)))
		db.Replicas = append(db.Replicas, r)
		return db, nil
	}
}

func (d *Daemon) applicationHandler(c echo.Context) error {
	rw, r := c.Response().Writer, c.Request()

	name := c.Param("name")

	d.applicationsLock.RLock()
	app, found := d.applications[name]
	d.applicationsLock.RUnlock()
	if !found {
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	}

	originalURL := r.URL
	// Remove name from path and route to correct port
	appURL := *r.URL
	appURL.Path = c.Param("_")
	appURL.Host = fmt.Sprintf("localhost:%d", app.port)
	appURL.Scheme = "http"
	r.URL = &appURL

	// Request.RequestURI can't be set in client requests.
	r.RequestURI = ""
	uri := r.RequestURI

	// TODO: Set to expected path for application to see
	// req.Header.Set("Host", "TODO")

	defer func() {
		// Restore the uri for use with the echo logger middleware
		r.RequestURI = uri
		r.URL = originalURL
	}()

	if err := app.newRequest(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "error starting process").Error())
	}
	defer app.endOfRequest()

	{
		resp, err := d.client.Do(r)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		_ = r.Body.Close()

		for k, v := range resp.Header {
			rw.Header()[k] = v
		}
		rw.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(rw, resp.Body)
		_ = resp.Body.Close()
	}
	return nil
}

func (d *Daemon) StopAllApplications() {
	d.applicationsLock.Lock()
	for _, app := range d.applications {
		app.stopProcess()
	}
	d.applicationsLock.Unlock()
}

func (d *Daemon) downloadDatabasesIfFound(dir, name string) (err error) {
	dbs, err := d.findDatabasesForApplication(name)
	if err != nil {
		return errors.Wrap(err, "error finding databases")
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
		generation, err := litestream.FindLatestGeneration(context.TODO(), replica.Client())
		if err != nil {
			return errors.Wrap(err, "finding latest generation")
		}
		targetIndex, err := litestream.FindMaxIndexByGeneration(context.TODO(), replica.Client(), generation)
		if err != nil {
			return errors.Wrap(err, "finding max index")
		}

		snapshotIndex, err := litestream.FindSnapshotForIndex(context.TODO(), replica.Client(), generation, targetIndex)
		if err != nil {
			return errors.Wrap(err, "finding snapshot for index")
		}
		if err := litestream.Restore(context.TODO(),
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
		Bucket:    aws.String("litestream"), // TODO
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

func (d *Daemon) validateAndAddApplication(name string, script []byte) (*Application, error) {
	// If two requests are validated simultaneously we won't catch them here,
	// but we'll error when adding them to the map later
	d.applicationsLock.RLock()
	_, found := d.applications[name]
	d.applicationsLock.RUnlock()
	if found {
		return nil, fmt.Errorf("an application with this name is already present on this host")
	}

	tmpDir, err := os.MkdirTemp(d.dataDirectory, "")
	if err != nil {
		return nil, err
	}
	fileName := filepath.Join(tmpDir, "index.ts")
	if err := os.WriteFile(fileName, script, 0600); err != nil {
		return nil, errors.Wrapf(err, "error creating file %q", fileName)
	}

	port, err := netx.GetFreePort()
	if err != nil {
		return nil, err
	}

	cmd, err := bunRun(tmpDir, port, nil)
	if err != nil {
		return nil, err
	}
	_ = cmd.Process.Kill()

	app := d.newApplication(name, tmpDir, port)
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
		if err := d.downloadDatabasesIfFound(tmpDir, name); err != nil {
			return nil, errors.Wrap(err, "error downloading database")
		}
	}
	app.DBs = dbs
	if err := app.dbDownladed(); err != nil {
		panic(err)
	}

	return app, nil
}

func (d *Daemon) Start(ctx context.Context) {
	if d.eg != nil {
		panic("Daemon has already started")
	}
	d.eg, ctx = errgroup.WithContext(ctx)

	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		// [GIN] 2023/01/23 - 21:00:26 | 200 |   30.208583ms |       127.0.0.1 | GET      "/max.hello/hi"
		Format: "${time_rfc3339} | ${status} | ${latency_human}\t | ${host} | ${method} | ${path} \n",
	}))
	api.RegisterHandlers(e, server{daemon: d})

	e.Any("/:name/*", d.applicationHandler)
	e.Any("/:name", d.applicationHandler)

	srv := http.Server{
		Handler:           e,
		ReadHeaderTimeout: time.Second * 15,
	}

	d.eg.Go(func() (err error) {
		d.listener, err = net.Listen("tcp", d.addr)
		if err != nil {
			return err
		}
		d.listenerWait.Done()
		err = srv.Serve(d.listener)
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	})
	d.eg.Go(func() error {
		<-ctx.Done()
		timeoutCtx, _ := context.WithTimeout(ctx, time.Minute) // TODO: vet this time
		if err := srv.Shutdown(timeoutCtx); err != nil {
			fmt.Printf("WARN: error shutting down http server: %v\n", err)
		}
		d.applicationsLock.Lock()
		defer d.applicationsLock.Unlock()

		for name, app := range d.applications {
			// TODO: will need to be parallel if we ever care about shutdown
			// speed with live applications
			if err := app.shutdown(); err != nil {
				fmt.Printf("WARN: error shutting down application %q: %v\n", name, err)
			}
		}
		return nil
	})

	// TODO: wait until server is live to exit?
}
