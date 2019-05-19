.EXPORT_ALL_VARIABLES:

APP_ENVIRONMENT = local-development
DATABASE_CONNECTION_STRING = postgres://postgres:password@localhost:5432/crashtested?sslmode=disable

run-api:
	go run ./cmd/api 

test:
	go run ./...
