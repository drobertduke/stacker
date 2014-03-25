package main

import (
	"fmt"
	"net/http"
	"github.com/hoisie/redis"
	"os"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi sthere, I love %s!", r.URL.Path[1:])

	var client redis.Client
	var redisHost = os.Getenv("REDIS_1_PORT_6379_TCP_ADDR")
	var redisPort = os.Getenv("REDIS_1_PORT_6379_TCP_PORT")
	client.Addr = redisHost + ":" + redisPort
	var key = "hello"
	client.Set(key, []byte("world"))
	val, _ := client.Get("hello")

	fmt.Fprintf(w, "Redis says: %s", string(val))
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
