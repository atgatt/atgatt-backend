package handlers

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

const MaxTimeToWait time.Duration = 10 * time.Second
const ApiBaseUrl string = "http://localhost:5001"

func WaitForApi() {
	isRunning := false
	timeWaited := time.Duration(0 * time.Millisecond)
	for !isRunning && timeWaited < MaxTimeToWait {
		fmt.Println("Waiting for server to come online...")
		resp, err := http.Get(ApiBaseUrl)
		if err == nil {
			isRunning = resp.StatusCode == http.StatusOK
		} else {
			fmt.Printf("Server returned an error: %s\n", err.Error())
		}
		if !isRunning {
			fmt.Println("Trying again after 200ms...")
			time.Sleep(200 * time.Millisecond)
			timeWaited += 200 * time.Millisecond
		}
	}
}

func TestMain(m *testing.M) {
	fmt.Println("Starting server in the background")
	server := Server{Port: ":5001", Name: "crashtested-api", Version: "integrationtests", BuildNumber: "1337"}
	go server.StartAndBlock()

	WaitForApi()
	fmt.Println("Server is running! Starting tests.")
	statusCode := m.Run()

	fmt.Println("Tests finished. Stopping server...")
	server.Stop()
	fmt.Println("Stopped server.")

	os.Exit(statusCode)
}
