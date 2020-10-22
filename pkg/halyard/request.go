package halyard

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type responseHolder struct {
	Err        error
	StatusCode int
	Body       []byte
}

type halyardValidateErrorResponse struct {
	ProblemSet struct {
		Problems []struct {
			Message string `json:"message"`
		} `json:"problems"`
	} `json:"problemSet"`
}

type halyardGenericErrorResponse struct {
	StatusLine string `json:"error,omitempty"`
	Message    string `json:"message,omitempty"`
}

func (s *Service) executeRequest(req *http.Request, ctx context.Context) responseHolder {
	req = req.WithContext(ctx)
	req.Header.Add("Accept-Encoding", "gzip")
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return responseHolder{Err: err}
	}
	defer resp.Body.Close()
	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		defer reader.Close()
	default:
		reader = resp.Body
	}

	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return responseHolder{Err: err, StatusCode: resp.StatusCode}
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return responseHolder{StatusCode: resp.StatusCode, Body: b}
	}
	return responseHolder{Body: b, StatusCode: resp.StatusCode}
}

func (hr *responseHolder) HasError() bool {
	return hr.Err != nil || hr.StatusCode < 200 || hr.StatusCode > 299
}

func (hr *responseHolder) Error() error {
	if hr.Err != nil {
		return hr.Err
	}
	if hr.StatusCode >= 200 && hr.StatusCode <= 299 {
		return nil
	}
	// try to get a friendly halyard error message from its response
	validateResp := &halyardValidateErrorResponse{}
	err := json.Unmarshal(hr.Body, &validateResp)
	if err != nil || len(validateResp.ProblemSet.Problems) == 0 {
		genResp := &halyardGenericErrorResponse{}
		err := json.Unmarshal(hr.Body, &genResp)
		if err != nil {
			return fmt.Errorf("got halyard response status %d, response: %s", hr.StatusCode, string(hr.Body))
		}
		return fmt.Errorf("got halyard response status %d, response: %s", hr.StatusCode, genResp.Message)
	}
	return fmt.Errorf(validateResp.ProblemSet.Problems[0].Message)
}
