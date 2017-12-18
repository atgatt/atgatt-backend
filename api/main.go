package main

import (
	"crashtested-backend/api/handlers"
)

func main() {
	s := &handlers.Server{Port: ":5000"}
	s.StartAndBlock()
}
