package main

import (
	"net/http"

	"github.com/maxmcd/steady/steady"
	"github.com/maxmcd/steady/steady/steadyrpc"
	"github.com/maxmcd/steady/web"
)

func main() {
	go func() {
		handler := steadyrpc.NewSteadyServer(
			steady.NewServer(
				steady.ServerOptions{},
				steady.OptionWithSqlite("./steady.sqlite")))
		panic(http.ListenAndServe(":8081", handler))
	}()

	s, err := web.NewServer(steadyrpc.NewSteadyProtobufClient("http://localhost:8081", http.DefaultClient))
	if err != nil {
		panic(err)
	}
	if err := s.Run(":8080"); err != nil {
		panic(err)
	}
}
