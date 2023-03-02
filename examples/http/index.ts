import { Database } from "bun:sqlite";

let db = new Database("db.sqlite");

db.exec(`create table if not exists user (
  id integer primary key autoincrement,
  email text
)`);

console.log(db.query("select * from user").all());

const port = process.env.PORT ?? 3000;
console.log("Listening on port " + port);

let insertUser = db.query("insert into user (email) values (?) returning *");

export default {
  port: port,
  async fetch(request: Request): Promise<Response> {
    console.log(`http: ${request.method} ${request.url}`)
    if (request.method !== "POST") {
      // return all users
      return new Response(JSON.stringify(db.query("select * from user").all()));
    }

    let req: { email: string } = await request.json();
    // console.log(req, request, JSON.stringify(request.headers));
    if (!req.email) return new Response("no", { status: 401 });
    // db.exec("insert into user (email) values (?)", req.email);

    // return new Response("");
    return new Response("All users: "+JSON.stringify(insertUser.all(req.email)[0]));
  },
};

