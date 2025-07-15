build:
	GOOS=linux GOARCH=amd64 go build -o main

logs:
	heroku logs --tail

.PHONY: build logs