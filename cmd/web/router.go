package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Helper function to attach all needed routes for web app
func (a *app) AttachRoutes(e *echo.Echo) {

	api := e.Group("/api")
	api.GET("/", a.HelloHandler)
	api.POST("/echo", a.EchoHandler)
}

// Simple handler example for GET /api
func (a *app) HelloHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, "Hello, World!")
}

// Simple handler example for POST /api/echo
// Will echo back the request body as a response
func (a *app) EchoHandler(c echo.Context) error {
	var body map[string]interface{}
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	return c.JSON(http.StatusOK, body)
}
