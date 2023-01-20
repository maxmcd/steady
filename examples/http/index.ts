import sqlite from "bun:sqlite";

let db: sqlite = (sqlite.default ? sqlite.default : sqlite).open("db.sql");
db.exec(
  "create table if not exists user (id integer primary key autoincrement, email text)"
);
console.log(db.query("select * from user").all());

const port = process.env.PORT ?? 3000;
console.log(`Listening on port ${port}`);

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
