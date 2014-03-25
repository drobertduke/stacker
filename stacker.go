package main

import (
	"fmt"
	"net/http"
	"github.com/hoisie/redis"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi sthere, I love %s!", r.URL.Path[1:])

	var client redis.Client
	var key = "hello"
	client.Set(key, []byte("world"))
	val, _ := client.Get("hello")

	fmt.Fprintf(w, string(val), r.URL.Path[1:])
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
