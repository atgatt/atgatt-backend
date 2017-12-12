package main

import (
	"crashtested-backend/api/server"
)

func main() {
	s := &server.Server{Port: ":5000"}
	s.StartAndBlock()
}
