package halyard

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type responseHolder struct {
	Err        error
	StatusCode int
	Body       []byte
}

type halyardErrorResponse struct {
	ProblemSet struct {
		Problems []struct {
			Message string `json:"message"`
		} `json:"problems"`
	} `json:"problemSet"`
}

func (s *Service) executeRequest(req *http.Request, ctx context.Context) responseHolder {
	req = req.WithContext(ctx)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return responseHolder{Err: err}
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
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
	resp := &halyardErrorResponse{}
	err := json.Unmarshal(hr.Body, &resp)
	if err != nil || len(resp.ProblemSet.Problems) == 0 {
		return fmt.Errorf("got halyard response status %d, response: %s", hr.StatusCode, string(hr.Body))
	}
	return fmt.Errorf(resp.ProblemSet.Problems[0].Message)
}
