package halyard

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/v1alpha2"
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"mime/multipart"
	"net/http"
	"strings"
)

type validationResponse []struct {
	Message  string `json:"message,omitempty"`
	Severity string `json:"severity,omitempty"`
	Location string `json:"location,omitempty"`
}

func (s *Service) Validate(ctx context.Context, spinsvc v1alpha2.SpinnakerServiceInterface, failFast bool, logger logr.Logger) error {
	req, err := s.buildValidationRequest(ctx, spinsvc, failFast)
	if err != nil {
		return err
	}
	resp := s.executeRequest(req, ctx)
	if resp.HasError() {
		return resp.Error()
	}
	return parseValidationResponse(resp.Body, logger)
}

func (s *Service) buildValidationRequest(ctx context.Context, spinsvc v1alpha2.SpinnakerServiceInterface, failFast bool) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// Add config
	cfg := spinsvc.GetSpinnakerConfig()
	b, err := yaml.Marshal(cfg.Config)
	if err != nil {
		return nil, err
	}
	if err = s.addPart(writer, "config", b); err != nil {
		return nil, err
	}
	// Add required files
	for k := range cfg.Files {
		if err := s.addPart(writer, k, cfg.GetFileContent(k)); err != nil {
			return nil, err
		}
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v1/validation/config?failFast=%t&skipValidators=%s", s.url, failFast, strings.Join(getValidationsToSkip(spinsvc.GetValidation()), ","))
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return req, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

func parseValidationResponse(d []byte, logger logr.Logger) error {
	resp := make(validationResponse, 0)
	if err := json.Unmarshal(d, &resp); err != nil {
		return errors.Wrap(err, "unable to read external validation response")
	}
	msgs := make([]string, 0)
	for _, v := range resp {
		if v.Severity == "FATAL" || v.Severity == "ERROR" {
			msgs = append(msgs, fmt.Sprintf("spinnakerConfig.config.%s: %s", v.Location, v.Message))
		} else {
			logger.Info(fmt.Sprintf("%s: %s at %s", v.Severity, v.Message, v.Location))
		}
	}
	if len(msgs) == 0 {
		return nil
	}
	return errors.New(strings.Join(msgs, ", "))
}
