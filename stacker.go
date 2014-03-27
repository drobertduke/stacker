package main

import (
	"net/http"
	"os"
	"log"
	"github.com/codegangsta/martini"
	"github.com/albrow/zoom"
	"github.com/garyburd/redigo/redis"
	"github.com/martini-contrib/binding"
)

type User struct {
	FullName string
	Username string
	zoom.DefaultData
}

type UserPost struct {
	FullName string `form:"fullName" json:"fullName" binding:"required"`
	Username string `form:"username" json:"username" binding:"required"`
}

func (up UserPost) Validate(errors *binding.Errors, req *http.Request) {
	lengthMax := 20
	if len(up.FullName) > lengthMax {
		errors.Fields["fullName"] = "Too long, should be less than " + string(lengthMax)
	}
	if len(up.Username) > lengthMax {
		errors.Fields["username"] = "Too long, should be less than " + string(lengthMax)
	}
}

func main() {
	var redisHost = os.Getenv("REDIS_1_PORT_6379_TCP_ADDR")
	var redisPort = os.Getenv("REDIS_1_PORT_6379_TCP_PORT")

	zoomConfig := &zoom.Configuration {
		Address: redisHost + ":" + redisPort,
		Network: "tcp",
	}
	zoom.Init(zoomConfig)

	c, _ := redis.Dial("tcp", redisHost + ":" + redisPort)

	if err := zoom.Register(&User{}); err != nil {
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

	m.Post("/users", binding.Bind(UserPost{}), func(params martini.Params) string {
		user := &User {
			FullName: params["fullName"],
			Username: params["username"],
		}
		if err := zoom.Save(user); err != nil {
			log.Fatal(err)
		}
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

