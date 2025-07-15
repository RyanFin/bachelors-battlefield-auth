build:
	GOOS=linux GOARCH=amd64 go build -o main

.PHONY: build