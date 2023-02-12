package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jmoiron/sqlx"
	_ "github.com/marcboeker/go-duckdb"
)

type Log struct {
	TS  int64  `db:"timestamp"`
	Log string `db:"log"`
}

func main() {
	db, err := sqlx.Open("duckdb", "./logs.db")
	if err != nil {
		panic(err)
	}

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS logs (
		timestamp BIGINT not null,
		log TEXT not null
	)`); err != nil {
		panic(err)
	}

	if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS logs_timestamp_idx ON logs (timestamp)`); err != nil {
		panic(err)
	}

	var count int

	// {

	// 	start := time.Now()
	// 	db.Exec(`select * from logs where log like '%MegaRAID%' order by "timestamp"`)
	// 	fmt.Println(time.Since(start))
	// 	start = time.Now()
	// 	db.Exec(`select * from logs where log like '%MegaRAID%' order by "timestamp"`)
	// 	fmt.Println(time.Since(start))
	// 	return
	// }

	db.Get(&count, "select count(*) from logs")
	fmt.Println(count)

	f, err := os.Open("Thunderbird.log")
	if err != nil {
		panic(err)
	}

	// var logs []Log
	scanner := bufio.NewScanner(f)

	c, err := os.Create("load.csv")
	if err != nil {
		panic(err)
	}
	start := time.Now()
	csvWriter := csv.NewWriter(c)
	if err := csvWriter.Write([]string{"timestamp", "log"}); err != nil {
		panic(err)
	}
	for scanner.Scan() {
		line := scanner.Text()
		if !utf8.Valid([]byte(line)) {
			continue
		}
		_, line, _ = strings.Cut(line, " ")
		date, log := line[:10], line[11:]
		i, err := strconv.Atoi(date)
		if err != nil {
			panic(err)
		}
		// logs = append(logs, Log{TS: time.Unix(int64(i), 0).UnixNano(), Log: log})

		if err := csvWriter.Write([]string{fmt.Sprint(time.Unix(int64(i), 0).UnixNano()), log}); err != nil {
			panic(err)
		}
	}
	csvWriter.Flush()
	fmt.Println("done csv", time.Since(start))
	start = time.Now()

	if _, err := db.Exec(`COPY logs FROM 'load.csv' (AUTO_DETECT TRUE);`); err != nil {
		panic(err)
	}
	fmt.Println("done copy", time.Since(start))

	db.Close()

}
