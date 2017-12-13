# crashtested-backend
The CrashTested backend services (API, background jobs, etc.)

## Build setup
- Run `dep ensure` to install all dependencies
- Run `go test ./...` to run all tests (integration and unit)
- Run `go build ./api` to build the API
- Run `./api` to run the API (listens on http://localhost:5000)

## Frameworks & Tools used
- Dep: Golang dependency management tool
- Echo: Simple HTTP server and routing framework, supports middleware