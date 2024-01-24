package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var app = fiber.New()

func TestRouterGetHelloWorld(t *testing.T) {

	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello World")
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello World", string(bytes))
}

func TestCtx(t *testing.T) {

	app.Get("/hello", func(ctx *fiber.Ctx) error {
		name := ctx.Query("name", "Guest")
		return ctx.SendString("Hello " + name)
	})

	request := httptest.NewRequest(http.MethodGet, "/hello?name=Ibra", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello Ibra", string(bytes))

	request = httptest.NewRequest(http.MethodGet, "/hello", nil)
	response, err = app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err = io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello Guest", string(bytes))
}

func TestHttpRequest(t *testing.T) {

	app.Get("/request", func(ctx *fiber.Ctx) error {
		first := ctx.Get("firstname")
		last := ctx.Cookies("lastname")
		return ctx.SendString("Hello " + first + " " + last)
	})

	request := httptest.NewRequest(http.MethodGet, "/request", nil)
	request.Header.Set("firstname", "Ibra")
	request.AddCookie(&http.Cookie{
		Name:  "lastname",
		Value: "Alfathar",
	})
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello Ibra Alfathar", string(bytes))
}

func TestRouteParameter(t *testing.T) {

	app.Get("/users/:userId/orders/:orderId", func(ctx *fiber.Ctx) error {
		userId := ctx.Params("userId")
		orderId := ctx.Params("orderId")

		return ctx.SendString("Get Order " + orderId + " From User " + userId)
	})

	request := httptest.NewRequest(http.MethodGet, "/users/ibra/orders/20", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Get Order 20 From User ibra", string(bytes))
}

func TestRequestForm(t *testing.T) {

	app.Post("/hello", func(ctx *fiber.Ctx) error {
		name := ctx.FormValue("name")
		return ctx.SendString("Hello " + name)
	})

	body := strings.NewReader("name=Ibra")
	request := httptest.NewRequest(http.MethodPost, "/hello", body)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Hello Ibra", string(bytes))
}

//go:embed source/contoh.txt
var contohFile []byte

func TestMultipartForm(t *testing.T) {

	app.Post("/upload", func(ctx *fiber.Ctx) error {
		file, err := ctx.FormFile("file")
		if err != nil {
			return err
		}

		err = ctx.SaveFile(file, "./target/"+file.Filename)
		if err != nil {
			return err
		}

		return ctx.SendString("Upload Success")

	})

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	file, err := writer.CreateFormFile("file", "contoh.txt")
	assert.Nil(t, err)
	_, err = file.Write(contohFile)
	if err != nil {
		panic(err)
	}

	writer.Close()

	request := httptest.NewRequest(http.MethodPost, "/upload", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	readAll, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Upload Success", string(readAll))
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func TestRequestBody(t *testing.T) {

	app.Post("/login", func(ctx *fiber.Ctx) error {
		body := ctx.Body()
		request := new(LoginRequest)
		err := json.Unmarshal(body, request)
		if err != nil {
			return err
		}
		return ctx.SendString("Login Success " + request.Username)
	})

	body := strings.NewReader(`{"username" : "Ibra", "password" : "test123"}`)
	request := httptest.NewRequest(http.MethodPost, "/login", body)
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Login Success Ibra", string(bytes))
}

type RegisterRequest struct {
	Username string `json:"username" xml:"username" form:"username"`
	Password string `json:"password" xml:"password" form:"password"`
	Name     string `json:"name" xml:"name" form:"name"`
}

func TestBodyParser(t *testing.T) {

	app.Post("/register", func(ctx *fiber.Ctx) error {
		request := new(RegisterRequest)
		err := ctx.BodyParser(request)
		if err != nil {
			return err
		}
		return ctx.SendString("Register Success " + request.Username)
	})

}

func TestBodyParserJSON(t *testing.T) {

	TestBodyParser(t)

	body := strings.NewReader(`{"username" : "Ibra", "password" : "test123", "name" : "ibra"}`)
	request := httptest.NewRequest(http.MethodPost, "/register", body)
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Register Success Ibra", string(bytes))
}

func TestBodyParserForm(t *testing.T) {

	TestBodyParser(t)

	body := strings.NewReader("username=Ibra&password=rahasia&name=ibra")
	request := httptest.NewRequest(http.MethodPost, "/register", body)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "Register Success Ibra", string(bytes))
}

func TestResponseJSON(t *testing.T) {

	app.Get("/user", func(ctx *fiber.Ctx) error {
		return ctx.JSON(fiber.Map{
			"username": "ibra",
			"name":     "ibra alfathar",
		})
	})

	request := httptest.NewRequest(http.MethodGet, "/user", nil)
	request.Header.Set("Accept", "application/json")
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, `{"name":"ibra alfathar","username":"ibra"}`, string(bytes))
}

func TestDownloadFile(t *testing.T) {

	app.Get("/download", func(ctx *fiber.Ctx) error {
		return ctx.Download("./source/contoh.txt", "contoh.txt")
	})

	request := httptest.NewRequest(http.MethodGet, "/download", nil)
	response, err := app.Test(request)
	assert.Nil(t, err)
	assert.Equal(t, 200, response.StatusCode)
	assert.Equal(t, `attachment; filename="contoh.txt"`, response.Header.Get("Content-Disposition"))

	bytes, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, "sample file for upload", string(bytes))
}
