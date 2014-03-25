package main

import (
	"net/http"
	"os"
	"github.com/hoisie/redis"
	"github.com/codegangsta/martini"
)

func main() {
	m := martini.Classic()

	m.Get("/", func() string {

		var client redis.Client
		var redisHost = os.Getenv("REDIS_1_PORT_6379_TCP_ADDR")
		var redisPort = os.Getenv("REDIS_1_PORT_6379_TCP_PORT")
		client.Addr = redisHost + ":" + redisPort
		var key = "hello"
		client.Set(key, []byte("world"))
		val, _ := client.Get("hello")

		return "martini " + string(val)
	});

	//m.Run()
	http.ListenAndServe(":8080", m)
}

