package main

import (
	"strconv"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/albrow/zoom"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/cors"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
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
	Priority    int
	Accepted    bool
	OwnerId     string
	zoom.DefaultData
}

type UserPost struct {
	FullName string `form:"FullName" json:"FullName"`
	Username string `form:"Username" json:"Username"`
	Id       string `form:"Id" json:"Id"`
}

type TaskPost struct {
	Title       string `form:"Title" json:"Title" binding:"required"`
	Description string `form:"Description" json:"Description:`
	OwnerId     string `form:"OwnerId" json:"OwnerId" binding:"required"`
}

func (up UserPost) Validate(errors *binding.Errors, req *http.Request) {
	lengthMax := 20
	if len(up.FullName) > lengthMax {
		errors.Fields["FullName"] = "Too long, should be less than " + string(lengthMax)
	}
	if len(up.Username) > lengthMax {
		errors.Fields["Username"] = "Too long, should be less than " + string(lengthMax)
	}
}

func (tp TaskPost) Validate(errors *binding.Errors, req *http.Request) {
	lengthMax := 50
	if len(tp.Title) > lengthMax {
		errors.Fields["Title"] = "Too long, should be less than " + string(lengthMax)
	}
	lengthMax = 1000
	if len(tp.Description) > lengthMax {
		errors.Fields["Description"] = "Too long, should be less than " + string(lengthMax)
	}

}

type JSendError struct {
	Message string `json:"message"`
}

func renderError(code int, err error, res http.ResponseWriter) string {
	jSend := &JSendError{err.Error()}
	j, err := json.Marshal(jSend)
	if err != nil {
		res.WriteHeader(500)
		return "Could not render error"
	}
	res.WriteHeader(code)
	return string(j)
}

func renderResponse(obj interface{}, res http.ResponseWriter) string {
	j, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		return renderError(500, err, res)
	}
	return string(j)
}

func getModelById(modelType reflect.Type, modelId string) (zoom.Model, error) {
	model, err := zoom.FindById(modelType.Name(), modelId)
	return model, err
}

func getModel(modelType reflect.Type, id string, res http.ResponseWriter) string {
	model, err := getModelById(modelType, id)
	if err != nil {
		return renderError(400, err, res)
	}
	return renderResponse(model, res)
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
		singleVal := val[0]
		if key == "Id" {
			continue
		}
		field := reflect.ValueOf(model).Elem().FieldByName(key)
		if !field.IsValid() {
			return renderError(400, errors.New("Field "+key+" is not valid"), res)
		}
		if !field.CanSet() {
			return renderError(400, errors.New("Field "+key+" cannot be set"), res)
		}
		if len(val) > 1 {
			field.Set(reflect.ValueOf(val))
		} else if field.Kind().String() == "string" {
			field.SetString(singleVal)
		} else if field.Kind().String() == "int" {
			intSingleVal, err := strconv.ParseInt(singleVal, 10, 64)
			if err != nil {
				return renderError(500, errors.New("Could not parse int"), res)
			}
			field.SetInt(intSingleVal)
		} else {
			return renderError(500, errors.New("Invalid field type"), res)
		}
	}
	if err := zoom.Save(model); err != nil {
		return renderError(400, err, res)
	}
	return renderResponse(model, res)
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

	store := sessions.NewCookieStore([]byte("secret123"))
	store.Options(sessions.Options{"/", "localhost", 0, false, false})
	m.Use(render.Renderer())
	m.Use(sessions.Sessions("my_session", store))
	m.Use(cors.Allow(&cors.Options{
		AllowOrigins:     []string{"http://127.0.0.1:9000"},
		AllowMethods:     []string{"PUT", "PATCH", "POST", "DELETE", "GET"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	m.Options("/tasks/**", func() {
	})
	m.Options("/tasks", func() {
	})
	m.Options("/users", func() {
	})

	// USERS

	m.Get("/users", func(res http.ResponseWriter, req *http.Request) string {
		results, err := zoom.NewQuery("User").Run()
		if err != nil {
			return renderError(400, err, res)
		}

		return renderResponse(results, res)
	})

	m.Get("/users/:userId", func(params martini.Params, res http.ResponseWriter, req *http.Request) string {
		return getModel(reflect.TypeOf(User{}), params["userId"], res)
	})

	m.Get("/users/:userId/tasks", func(params martini.Params, res http.ResponseWriter, req *http.Request) string {
		model, err := getModelById(reflect.TypeOf(User{}), params["userId"])
		if err != nil {
			return renderError(400, err, res)
		}
		user := model.(*User)
		modelNames := make([]string, len(user.TaskIds))
		for i := range user.TaskIds {
			modelNames[i] = "Task"
		}
		tasks, err := zoom.MFindById(modelNames, user.TaskIds)
		if err != nil {
			return renderError(400, err, res)
		}
		return renderResponse(tasks, res)
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
			uri := req.URL.Scheme + req.URL.Host + req.URL.Path + "/" + user.Id
			res.Header().Add("Location", uri)
			res.WriteHeader(201)
			return renderResponse(user, res)
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

		return renderResponse(results, res)
	})

	m.Get("/tasks/:taskId", func(params martini.Params, res http.ResponseWriter, req *http.Request) string {
		return getModel(reflect.TypeOf(Task{}), params["taskId"], res)
	})

	m.Post(
		"/tasks",
		binding.Bind(TaskPost{}),
		binding.ErrorHandler,
		func(taskPost TaskPost, res http.ResponseWriter, req *http.Request) string {
			model, err := getModelById(reflect.TypeOf(User{}), taskPost.OwnerId)
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
			user := model.(*User)
			user.TaskIds = append(user.TaskIds, task.Id)
			if err := zoom.Save(user); err != nil {
				return renderError(400, err, res)
			}
			uri := req.URL.Scheme + req.URL.Host + req.URL.Path + "/" + task.Id
			res.WriteHeader(201)
			res.Header().Add("Location", uri)
			return renderResponse(task, res)
		})

	m.Put("/tasks/:taskId", func(params martini.Params, res http.ResponseWriter, req *http.Request) string {
		return putModel(reflect.TypeOf(Task{}), params["taskId"], res, req)
	})

	m.Delete("/tasks/:taskId", func(params martini.Params, res http.ResponseWriter) string {
		zoom.DeleteById("Task", params["taskId"])
		res.WriteHeader(204)
		return ""
	})

	http.ListenAndServe(":8081", m)
}
