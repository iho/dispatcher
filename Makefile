gen:
	 mockgen -source=client/api_client.go > mock_client/api_client.go

install:
	 go install go.uber.org/mock/mockgen@latest

test:
	go test -v ./...

run:
	go run main.go
