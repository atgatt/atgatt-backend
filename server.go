package main

import (
	"net/http"

	"github.com/labstack/echo"
)

func main() {
	echoPtr := echo.New()
	echoPtr.GET("/", func(context echo.Context) error {
		return context.String(http.StatusOK, "Hello, World!")
	})
	echoPtr.Logger.Fatal(echoPtr.Start(":5000"))
}
