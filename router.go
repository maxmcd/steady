package steady

import (
	"net/http"
	"net/url"
)

func Handler(w http.ResponseWriter, r *http.Request) {

}

type Request struct {
	URL    url.URL
	Method string
	Body   []byte
	Header http.Header
}

type Response struct {
	Status int
	Body   int
	Header http.Header
}
