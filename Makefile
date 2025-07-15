build:
	GOOS=linux GOARCH=amd64 go build -o main

logs:
	heroku logs --tail

push-to-heroku: build
	git push heroku main

.PHONY: build logs