package main

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/evalphobia/github-pr-checker/prchecker"
)

func main() {
	handler, err := prchecker.New()
	if err != nil {
		panic(err)
	}

	lambda.Start(func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		body := ioutil.NopCloser(strings.NewReader(request.Body))
		header := make(http.Header)
		for key, val := range request.Headers {
			header.Set(key, val)
		}

		err := handler.HandleRequest(&http.Request{
			Method: request.HTTPMethod,
			Header: header,
			Body:   body,
		})
		if err != nil {
			return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 500}, err
		}
		return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 200}, nil
	})
}
