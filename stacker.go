package main

import (
	"encoding/json"
	"errors"
	"github.com/albrow/zoom"
	"github.com/codegangsta/martini"
	"github.com/garyburd/redigo/redis"
	"github.com/martini-contrib/binding"
	"log"
	"net/http"
	"os"
)

type User struct {
	FullName string
	Username string
	Tasks    []*Task
	zoom.DefaultData
}

type Task struct {
	Title       string
	Description string
	OwnerId     string
	zoom.DefaultData
}

type UserPost struct {
	FullName string `form:"fullName" json:"fullName" binding:"required"`
	Username string `form:"username" json:"username" binding:"required"`
}

type TaskPost struct {
	Title       string `form:"title" json:"title" binding:"required"`
	Description string `form:"description" json:"description:`
	OwnerId     string `form:"ownerId" json:"ownerId" binding:"required"`
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

type JSendResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

type JSendError struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

const (
	JSendStatusSuccess = "success"
	JSendStatusError   = "error"
)

func renderError(code int, err error, res http.ResponseWriter) string {
	jSend := &JSendError{JSendStatusError, err.Error()}
	j, err := json.Marshal(jSend)
	if err != nil {
		res.WriteHeader(500)
		return "Could not render error"
	}
	res.WriteHeader(code)
	return string(j)
}

func renderResponse(obj interface{}, objType string, res http.ResponseWriter) string {
	jSend := &JSendResponse{JSendStatusSuccess, map[string]interface{}{objType: obj}}
	j, err := json.Marshal(jSend)
	if err != nil {
		return renderError(500, err, res)
	}
	return string(j)
}

func getUserById(userId string) (zoom.Model, error) {
	return zoom.FindById("User", userId)
}

func main() {
	var redisHost = os.Getenv("REDIS_1_PORT_6379_TCP_ADDR")
	var redisPort = os.Getenv("REDIS_1_PORT_6379_TCP_PORT")

	zoomConfig := &zoom.Configuration{
		Address: redisHost + ":" + redisPort,
		Network: "tcp",
	}

	zoom.Init(zoomConfig)

	c, _ := redis.Dial("tcp", redisHost+":"+redisPort)

	if err := zoom.Register(&User{}); err != nil {
		log.Fatal(err)
	}

	if err := zoom.Register(&Task{}); err != nil {
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
	})

	// API

	// USERS

	m.Get("/users", func(res http.ResponseWriter, req *http.Request) string {
		results, err := zoom.NewQuery("User").Run()
		if err != nil {
			return renderError(400, err, res)
		}

		return renderResponse(results, "users", res)
	})

	m.Get("/users/:userId", func(params martini.Params, res http.ResponseWriter, req *http.Request) string {
		result, err := getUserById(params["userId"])
		if err != nil {
			return renderError(400, err, res)
		}
		user, ok := result.(*User)
		if !ok {
			return renderError(400, errors.New("No user found"), res)
		}
		return renderResponse(user, "user", res)
	})

	m.Get("/users/:userId/tasks", func(params martini.Params, res http.ResponseWriter, req *http.Request) string {
		return "TASKS FOR USER " + params["userId"]
	})

	m.Post(
		"/users",
		binding.Bind(UserPost{}),
		binding.ErrorHandler,
		func(userPost UserPost, res http.ResponseWriter, req *http.Request) string {

			user := &User{
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

	m.Get("/tasks", func(res http.ResponseWriter, req *http.Request) string {
		results, err := zoom.NewQuery("Task").Run()
		if err != nil {
			return renderError(400, err, res)
		}

		return renderResponse(results, "tasks", res)
	})

	m.Post(
		"/tasks",
		binding.Bind(TaskPost{}),
		binding.ErrorHandler,
		func(taskPost TaskPost, res http.ResponseWriter, req *http.Request) string {
			userResult, err := getUserById(taskPost.OwnerId)
			if err != nil {
				return renderError(400, err, res)
			}
			user := userResult.(*User)
			task := &Task{
				Title:       taskPost.Title,
				Description: taskPost.Description,
				OwnerId:     taskPost.OwnerId,
			}
			user.Tasks = append(user.Tasks, task)
			if err := zoom.Save(task); err != nil {
				return renderError(400, err, res)
			}
			if err := zoom.Save(user); err != nil {
				return renderError(400, err, res)
			}
			return "POSTED TASK"
		})

	m.Put("/tasks/:taskId", func(params martini.Params) string {
		return "PUT TASK" + params["taskId"]
	})

	http.ListenAndServe(":8080", m)
}
