package main

import (
	"net/http"

	"github.com/maxmcd/steady/steady"
	"github.com/maxmcd/steady/steady/steadyrpc"
	"github.com/maxmcd/steady/web"
)

func main() {
	steadyHandler :=
		steady.NewServer(
			steady.ServerOptions{},
			steady.OptionWithSqlite("./steady.sqlite"))

	webHandler, err := web.NewServer(steadyrpc.NewSteadyProtobufClient("http://localhost:8080", http.DefaultClient))
	if err != nil {
		panic(err)
	}

	panic(
		http.ListenAndServe(":8080", web.WebAndSteadyHandler(steadyHandler, webHandler)),
	)
}
