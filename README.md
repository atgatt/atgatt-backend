# crashtested-backend
Monorepo for all of the CrashTested backend services (API, background jobs, etc.)

[![CircleCI](https://circleci.com/gh/bakatz/crashtested-backend.svg?style=svg&circle-token=0aafe6b739c14e33dd07db99920ee7a82aa4d30b)](https://circleci.com/gh/bakatz/crashtested-backend)

## First-time setup
1. Clone this repository to your local machine in some directory i.e. `~/dev/crashtested-backend`
1. Install VS Code via https://code.visualstudio.com/ 
1. In VS Code, install the `Go` extension to get code completion and other nice things
1. Install Go 1.11.x via https://golang.org/dl/
1. Install Postgres 10.6 via https://www.postgresql.org/download/
1. Install sql-migrate by running `go get -v github.com/rubenv/sql-migrate/...`
1. Optionally, install Air for live reload support via https://github.com/cosmtrek/air#installation

NOTE: You don't need to do anything to install dependencies. This project relies on the new Go Modules feature, which means that when you `go build` the API/background worker for the first time, Go will automatically install all deps.

## Common tasks
- Run `go run ./cmd/api` to run the API (listens on http://localhost:5000)
- Run `go test ./...` to run all tests (integration and unit)
- Run `sql-migrate new <name-of-migration>` to generate a new migration
- Run `sql-migrate up` to run migrations
- Run `go build -o ./crashtested-api ./cmd/api` to build the API to a self-contained binary
- Run `go build -o ./crashtested-worker ./cmd/worker` to build the background worker to a self-contained binary
- If you have Air, type `air` (or `air -c .air.windows.conf` if you're on Windows) to run a live reload server. 

## Important folders and files
- `api` - controllers and request handling logic
- `worker` - background jobs
- `cmd` - main.go files i.e. entrypoints for `api` and `worker`
- `go.mod` - all of the dependencies for the project

## Deployment
### Environments
- Staging: 
    - API Healthcheck URL: https://api-staging.crashtested.co/
    - API Elastic Beanstalk Environment Name: `api-staging`
    - Worker Elastic Beanstalk Environment Name: `worker-staging`
- Production: 
    - API Healthcheck URL: https://api.crashtested.co/
    - API Elastic Beanstalk Environment Name: `api-prod`
    - Worker Elastic Beanstalk Environment Name: `worker-prod`

### How to deploy
- Staging: Just merge your feature branch to master. After it gets merged, it will automatically get deployed to staging.
- Production: Select the commit you want to promote to production via https://circleci.com/gh/bakatz/workflows/crashtested-backend, then click on the `hold` step and click `Approve` to promote to production
- Any failures for either of the above steps will be sent to #build-notifications in Slack

### Notes
- CrashTested is hosted on AWS using Elastic Beanstalk using a web role and a worker role. A `Procfile` for each role (two separate files found in ./api and ./worker) controls how the service is started once it's deployed to EB.
- Elastic Beanstalk expects the API to run on port 5000 by default and the worker to run on port 80 by default, so do not change `server.go` to have it point to a different port.