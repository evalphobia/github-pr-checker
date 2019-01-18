package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/evalphobia/github-pr-checker/prchecker"
)

const (
	envHTTPPort     = "GITHUB_PR_HTTP_PORT"
	defaultHTTPPort = 3000
)

func main() {
	handler, err := prchecker.New()
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := handler.HandleRequest(r)
		if err != nil {
			fmt.Printf("[ERROR] %+v\n", err)
		}
		fmt.Fprintf(w, "{}")
	})

	if err := http.ListenAndServe(fmt.Sprintf(":%s", getHTTPPort()), nil); err != nil {
		panic(err)
	}
}

func getHTTPPort() string {
	port := os.Getenv(envHTTPPort)
	switch {
	case port != "":
		return port
	}
	return strconv.Itoa(defaultHTTPPort)
}
