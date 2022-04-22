package main

import (
	"io/ioutil"
	"net/http"
)

func main() {
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	request := duplicateRequest(r)
	body := requestToRealUpstream(request)
	w.Write(body)
}

func duplicateRequest(r *http.Request) *http.Request {
	// Get original request
	body := r.Body
	method := r.Method

	// Get original url to substitute Github
	githubURL := "https://github.com"
	phishingURL := githubURL + r.URL.Path + "?" + r.URL.RawQuery

	// create new http request
	request, err := http.NewRequest(method, phishingURL, body)
	if err != nil {
		panic(err)
	}
	return request
}

func requestToRealUpstream(req *http.Request) []byte {
	// Create an instance of HttpClient
	client := http.Client{}

	// Send a request to real upstream and receive the response
	response, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	// Convert the response body to byte
	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	response.Body.Close()

	return respBody
}
