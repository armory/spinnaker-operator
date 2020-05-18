package util

import (
	"context"
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

func (s *HttpService) Request(ctx context.Context, method HttpMethod, url string, requestParams map[string]string, headers map[string]string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(string(method), url, body)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

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

func (s *HttpService) Execute(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)
	client := &http.Client{}
	return client.Do(req)
}

func (s *HttpService) ParseResponseBody(body io.ReadCloser) ([]byte, error) {
	defer body.Close()
	f, err := ioutil.ReadAll(body)

	if err != nil {
		return nil, err
	}

	return f, nil
}
