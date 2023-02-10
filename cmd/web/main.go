package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/handlers"
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
		http.ListenAndServe(":8080",
			handlers.LoggingHandler(os.Stdout,
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if strings.HasPrefix(r.URL.Path, "/twirp") {
						steadyHandler.ServeHTTP(w, r)
					} else {
						webHandler.ServeHTTP(w, r)
					}
				}),
			),
		),
	)
}
