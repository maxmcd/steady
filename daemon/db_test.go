package daemon

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/benbjohnson/litestream"
	"github.com/benbjohnson/litestream/s3"
	"github.com/stretchr/testify/assert"
)

var dbServer = `
import sqlite from "bun:sqlite";

let db: sqlite = (sqlite.default ? sqlite.default : sqlite).open("db.sql");
db.exec(
  "create table if not exists user (id integer primary key autoincrement, email text)"
);
console.log(db.query("select * from user").all());

const port = process.env.PORT ?? 3000;
console.log("Listening on port "+port);

export default {
  port,
  async fetch(request: Request): Promise<Response> {
    if (request.method === "POST") {
      let req: { email: string } = await request.json();
      if (!req.email) return new Response("no");
      return new Response(
        JSON.stringify(
          db
            .query("insert into user (email) values (?) returning *")
            .all(req.email)
        )
      );
    }
    return new Response(JSON.stringify(db.query("select * from user").all()));
  },
};
`

func TestLitestream(t *testing.T) {
	d := NewDaemon(t.TempDir(), 8080)
	ctx, cancel := context.WithCancel(context.Background())
	d.Start(ctx)

	minIOServer := NewMinioServer(t)
	fmt.Println(minIOServer.Address)

	app, err := d.validateAndAddApplication("max.db", []byte(dbServer))
	if err != nil {
		t.Fatal(err)
	}
	_ = app

	createRecordRequest := func() {
		resp, err := http.Post("http://localhost:8080/max.db/", "application/json", bytes.NewBuffer([]byte(`{"email":"lite"}`)))
		if err != nil {
			t.Fatal(err)
		}
		b, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("%q\n", string(b))
		assert.Equal(t, http.StatusOK, resp.StatusCode)

	}

	createRecordRequest()

	server := litestream.NewServer()
	if err := server.Open(); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(app.dir, "db.sql")
	if err := server.Watch(dbPath, func(path string) (_ *litestream.DB, err error) {
		db := litestream.NewDB(path)

		client := s3.NewReplicaClient()
		client.AccessKeyID = minIOServer.Username
		client.SecretAccessKey = minIOServer.Password
		client.Bucket = minIOServer.BucketName
		client.Path = "bar"
		client.Endpoint = "http://" + minIOServer.Address
		client.SkipVerify = true
		client.ForcePathStyle = true
		r := litestream.NewReplica(db, path, client)
		db.Replicas = append(db.Replicas, r)
		return db, nil
	}); err != nil {
		t.Fatal(err)
	}

	createRecordRequest()

	time.Sleep(time.Second)

	cancel()
	if err := d.Wait(); err != nil {
		t.Fatal(err)
	}
}
