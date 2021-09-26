package main

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"net/http"
	_url "net/url"
	"strings"
)

func HttpPost(url string, body map[string]string, params map[string]string, headers map[string]string) (*http.Response, error) {
	//add post body
	//var bodyJson []byte
	var req *http.Request
	var data = _url.Values{}
	if body != nil {
		for key, val := range body {
			data.Add(key, val)
		}
	}
	_body := strings.NewReader(data.Encode())
	req, err := http.NewRequest("POST", url, _body)
	if err != nil {
		log.Println(err)
		return nil, errors.New("new request is fail: %v \n")
	}

	//add params
	q := req.URL.Query()
	if params != nil {
		for key, val := range params {
			q.Add(key, val)
		}
		req.URL.RawQuery = q.Encode()
	}
	//add headers
	if headers != nil {
		for key, val := range headers {
			req.Header.Add(key, val)
		}
	}
	//requestDump, err := httputil.DumpRequest(req, true)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println(string(requestDump))
	//http client
	client := &http.Client{}
	//log.Printf("Go POST URL : %s \n", req.URL.String())
	return client.Do(req)
}

func HttpGet(url string, params map[string]string, headers map[string]string) (*http.Response, error) {
	//new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
		return nil, errors.New("new request is fail ")
	}
	//add params
	q := req.URL.Query()
	if params != nil {
		for key, val := range params {
			q.Add(key, val)
		}
		req.URL.RawQuery = q.Encode()
	}
	//add headers
	if headers != nil {
		for key, val := range headers {
			req.Header.Add(key, val)
		}
	}
	//http client
	client := &http.Client{}
	log.Printf("Go GET URL : %s \n", req.URL.String())
	return client.Do(req)
}
