test:
	go test ./...

start:
	docker-compose up

createEvents:
	go run api/cmd/main.go