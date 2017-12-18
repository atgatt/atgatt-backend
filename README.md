# crashtested-backend
The CrashTested backend services (API, background jobs, etc.)

## Build setup
- Run `dep ensure` to install all dependencies
- Run `go test ./...` to run all tests (integration and unit)
- Run `go build -o ./crashtested-api ./api` to build the API
- Run `./crashtested-api` to run the API (listens on http://localhost:5000)

## Notes about deployment
- CrashTested is hosted on AWS using Elastic Beanstalk. `Procfile` controls how the app is started once it's deployed to EB.
- Elastic Beanstalk expects the app to run on port 5000 by default, so do not change `server.go` to have it point to a different port.

## Frameworks & Tools used
- Dep: Golang dependency management tool
- Echo: Simple HTTP and routing framework, supports middleware
- Air: Live reload for Go - makes local development easy by automatically running a certain command defined in `.air.conf` when files in the workspace change
