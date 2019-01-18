.PHONY: init dep build deploy

init:
	go get -u github.com/golang/dep/cmd/dep
	npm install -g serverless

dep:
	dep ensure -v

build:
	GOOS=linux go build -o bin/serverless cmd/serverless/main.go

deploy: dep build
	sls deploy -v
