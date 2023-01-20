package main

import (
	"fmt"

	"github.com/benbjohnson/litestream"
)

func main() {

	server := litestream.NewServer()
	fmt.Println(server)
}
