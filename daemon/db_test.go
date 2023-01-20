package daemon

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/benbjohnson/litestream"
	"github.com/benbjohnson/litestream/s3"
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

	app, err := d.validateAndAddApplication("max.db", []byte(dbServer))
	if err != nil {
		t.Fatal(err)
	}
	_ = app

	server := litestream.NewServer()
	if err := server.Open(); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(d.dataDirectory, "db.sql")
	if err := server.Watch(dbPath, func(path string) (_ *litestream.DB, err error) {
		db := litestream.NewDB(path)

		client := s3.NewReplicaClient()
		client.AccessKeyID = ""
		client.SecretAccessKey = ""
		client.Bucket = "foo"
		client.Path = "bar"
		client.Endpoint = "http://" + minIOServer.Address
		client.SkipVerify = true
		r := litestream.NewReplica(db, path, client)
		db.Replicas = append(db.Replicas, r)
		return db, nil
	}); err != nil {
		t.Fatal(err)
	}
	cancel()
	if err := d.Wait(); err != nil {
		t.Fatal(err)
	}
}
