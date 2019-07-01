.EXPORT_ALL_VARIABLES:

APP_ENVIRONMENT = local-development
DATABASE_CONNECTION_STRING = postgres://postgres:password@localhost:5432/crashtested?sslmode=disable
AUTH0_DOMAIN = crashtested-staging.auth0.com

run-api:
	go run ./cmd/api

test:
	go test ./...
