package util

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

type HttpService struct {
}

type HttpMethod string

const (
	GET    HttpMethod = "GET"
	POST   HttpMethod = "POST"
	PUT    HttpMethod = "PUT"
	DELETE HttpMethod = "DELETE"
)

func (s *HttpService) Request(method HttpMethod, url string, requestParams map[string]string, headers map[string]string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(string(method), url, body)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(context.TODO())

	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	if requestParams != nil {
		q := req.URL.Query()
		for k, v := range requestParams {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	return req, nil
}

func (s *HttpService) Execute(req *http.Request) (*http.Response, error) {
	req = req.WithContext(context.TODO())
	client := &http.Client{}
	return client.Do(req)
}

func (s *HttpService) ParseResponseBody(body io.ReadCloser) (map[string]interface{}, error) {
	defer body.Close()
	f, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}

	response := make(map[string]interface{})
	if len(f) != 0 {
		if err := json.Unmarshal(f, &response); err != nil {
			return nil, err
		}
	}

	return response, nil
}
