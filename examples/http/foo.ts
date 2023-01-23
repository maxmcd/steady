import sqlite from "bun:sqlite";

let db: sqlite = (sqlite.default ? sqlite.default : sqlite).open("ohno.sql");

db.exec(`create table if not exists user (
  id integer primary key autoincrement,
  email text
)`);
console.log(db.query("select * from user").all());

let insertUser = db.query("insert into user (email) values (?) returning *");

insertUser.all("hi");
insertUser.all("hi");
insertUser.all("hi");
insertUser.all("hi");
