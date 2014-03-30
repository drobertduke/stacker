package main

import (
	"net/http"
	"os"
	"log"
	"github.com/codegangsta/martini"
	"github.com/albrow/zoom"
	"github.com/garyburd/redigo/redis"
	"github.com/martini-contrib/binding"
	"reflect"
	"encoding/json"
	"errors"
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

type ErrorResponse struct {
	Message string
}

func renderError(code int, err error, res http.ResponseWriter) string {
	msg := &ErrorResponse{err.Error()}
	j, err := json.Marshal(msg)
	if err != nil {
		res.WriteHeader(500)
		return "Could not return error"
	}
	res.WriteHeader(code)
	return string(j)
}

function renderSuccess()

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

	m.Get("/users", func(res http.ResponseWriter, req *http.Request) string {
		results, err := zoom.NewQuery("User").Run()
		if err != nil {
			return renderError(400, err, res)
		}
		users := reflect.ValueOf(results)
		resp := ""
		for i := 0; i < users.Len(); i++ {
			user := users.Index(i).Interface().(*User)
			j, err := json.Marshal(user)
			if err != nil {
				return renderError(500, err, res)
			}
			resp = resp + string(j)
		}
		return resp
	})

	m.Get("/users/:userId", func(params martini.Params, res http.ResponseWriter, req *http.Request) string {
		result, err := zoom.FindById("User", params["userId"])
		if err != nil {
			return renderError(400, err, res)
		}
		user, ok := result.(*User)
		if !ok {
			return renderError(500, errors.New("Could not case to User"), res)
		}
		j, err := json.Marshal(user)
		if err != nil {
			return renderError(500, err, res)
		}
		return string(j)
	})

	m.Get("/users/:userId/tasks", func(params martini.Params, res http.ResponseWriter, req *http.Request) string {
		return "TASKS FOR USER " + params["userId"]
	})

	m.Post(
		"/users",
		binding.Bind(UserPost{}),
		binding.ErrorHandler,
		func(userPost UserPost, res http.ResponseWriter, req *http.Request) string {

		user := &User {
			FullName: userPost.FullName,
			Username: userPost.Username,
		}
		if err := zoom.Save(user); err != nil {
			return renderError(400, err, res)
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

