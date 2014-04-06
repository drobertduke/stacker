package main

import (
	"strings"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/albrow/zoom"
	"github.com/codegangsta/martini"
	"github.com/martini-contrib/binding"
	"log"
	"net/http"
	"os"
	"reflect"
)

type User struct {
	FullName string
	Username string
	TaskIds  []string
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

func getUserById(userId string) (*User, error) {
	user := &User{}
	err := zoom.ScanById(userId, user)
	return user, err
}

func getTaskById(taskId string) (*Task, error) {
	task := &Task{}
	err := zoom.ScanById(taskId, task)
	return task, err
}

func putModel(modelType reflect.Type, id string, res http.ResponseWriter, req *http.Request) string {

	if req.ParseForm() != nil {
		return renderError(400, errors.New("Could not parse form data"), res)
	}
	if len(req.Form) == 0 {
		return renderError(400, errors.New("Please provide fields to update"), res)
	}
	result, err := zoom.FindById(modelType.Name(), id)
	if err != nil {
		return renderError(400, err, res)
	}
	model := result.(zoom.Model)
	for key, val := range req.Form {
		if len(val) > 1 {
			return renderError(400, errors.New("Cannot set field " + key + " to an array"), res)
		}
		singleVal := val[0]
		if key == "Id" { continue }
		field := reflect.ValueOf(model).Elem().FieldByName(key)
		if !field.IsValid() {
			return renderError(400, errors.New("Field " + key + " is not valid"), res)
		}
		if !field.CanSet() {
			return renderError(400, errors.New("Field " + key + " cannot be set"), res)
		}
		field.SetString(singleVal)
	}
	if err := zoom.Save(model); err != nil {
		return renderError(400, err, res)
	}
	return renderResponse(model, strings.ToLower(modelType.Name()), res)
}

func main() {
	fmt.Println(os.Getenv("REDIS_1_PORT_6379_TCP_ADDR"))
	var redisHost = os.Getenv("REDIS_1_PORT_6379_TCP_ADDR")
	var redisPort = os.Getenv("REDIS_1_PORT_6379_TCP_PORT")

	zoomConfig := &zoom.Configuration{
		Address: redisHost + ":" + redisPort,
		Network: "tcp",
	}

	zoom.Init(zoomConfig)

	if err := zoom.Register(&User{}); err != nil {
		log.Fatal(err)
	}

	if err := zoom.Register(&Task{}); err != nil {
		log.Fatal(err)
	}
	m := martini.Classic()

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
		user, err := getUserById(params["userId"])
		if err != nil {
			return renderError(400, err, res)
		}
		return renderResponse(user, "user", res)
	})

	m.Get("/users/:userId/tasks", func(params martini.Params, res http.ResponseWriter, req *http.Request) string {
		user, err := getUserById(params["userId"])
		if err != nil {
			return renderError(400, err, res)
		}
		modelNames := make([]string, len(user.TaskIds))
		for i := range user.TaskIds {
			modelNames[i] = "Task"
		}
		tasks, err := zoom.MFindById(modelNames, user.TaskIds)
		return renderResponse(tasks, "tasks", res)
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
			return renderResponse(user, "user", res)
		})

	m.Put("/users/:userId", func(params martini.Params, res http.ResponseWriter, req *http.Request) string {
		return putModel(reflect.TypeOf(User{}), params["userId"], res, req)
	})

	// TASKS

	m.Get("/tasks", func(res http.ResponseWriter, req *http.Request) string {
		results, err := zoom.NewQuery("Task").Run()
		if err != nil {
			return renderError(400, err, res)
		}

		return renderResponse(results, "tasks", res)
	})

	m.Get("/tasks/:taskId", func(params martini.Params, res http.ResponseWriter, req *http.Request) string {
		task, err := getTaskById(params["taskId"])
		if err != nil {
			return renderError(400, err, res)
		}
		return renderResponse(task, "task", res)
	})

	m.Post(
		"/tasks",
		binding.Bind(TaskPost{}),
		binding.ErrorHandler,
		func(taskPost TaskPost, res http.ResponseWriter, req *http.Request) string {
			user, err := getUserById(taskPost.OwnerId)
			if err != nil {
				return renderError(400, err, res)
			}
			task := &Task{
				Title:       taskPost.Title,
				Description: taskPost.Description,
				OwnerId:     taskPost.OwnerId,
			}
			if err := zoom.Save(task); err != nil {
				return renderError(400, err, res)
			}
			user.TaskIds = append(user.TaskIds, task.Id)
			if err := zoom.Save(user); err != nil {
				return renderError(400, err, res)
			}
			return renderResponse(task, "task", res)
		})

	m.Put("/tasks/:taskId", func(params martini.Params, res http.ResponseWriter, req *http.Request) string {
		return putModel(reflect.TypeOf(Task{}), params["taskId"], res, req)
	})

	m.Delete("/tasks/:taskId", func(params martini.Params) string {
		zoom.DeleteById("Task", params["taskId"])
		return "DELETED"
	})

	http.ListenAndServe(":8080", m)
}
