package main

import (
	"crashtested-backend/api/configuration"
	"crashtested-backend/api/handlers"
)

func main() {
	s := &handlers.Server{Port: ":5000", Name: "crashtested-api", Version: "1.0.0", BuildNumber: "{LOCAL-DEV-BUILD}", Configuration: configuration.GetDefaultConfiguration()}
	s.StartAndBlock()
}
