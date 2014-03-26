package main

import (
	"net/http"
	"os"
	"log"
	"github.com/codegangsta/martini"
	"github.com/garyburd/redigo/redis"
)

func main() {
	var redisHost = os.Getenv("REDIS_1_PORT_6379_TCP_ADDR")
	var redisPort = os.Getenv("REDIS_1_PORT_6379_TCP_PORT")
	c, err := redis.Dial("tcp", redisHost + ":" + redisPort)
	if err != nil {
		log.Fatal(err)
	}

	m := martini.Classic()

	c.Do("SET", "hello", "world")

	m.Get("/", func() string {

		val, _ := c.Do("GET", "hello")
		bytes, ok := val.([]byte)
		if !ok {
			log.Fatal("Whoops")
		}
		return "martini " + string(bytes)
	});

	// API

	// USERS

	m.Get("/users", func() string {
		return "list of users"
	})

	m.Get("/users/:userId", func(params martini.Params) string {
		return "DETAIL FOR USER " + params["userId"]
	})

	m.Get("/users/:userId/tasks", func(params martini.Params) string {
		return "TASKS FOR USER " + params["userId"]
	})

	m.Post("/users", func(params martini.Params) string {
		return "POSTED USER"
	})

	m.Put("/users/:userId", func(params martini.Params) string {
		return "PUT USER " + params["userId"]
	})

	// TASKS

	m.Get("/tasks", func() string {
		return "list of tasks"
	})

	m.Post("/tasks", func() string {
		return "POSTED TASK"
	})

	m.Put("/tasks/:taskId", func(params martini.Params) string {
		return "PUT TASK" + params["taskId"]
	})

	http.ListenAndServe(":8080", m)
}

