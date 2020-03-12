package halyard

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/armory/spinnaker-operator/pkg/secrets"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path"
	"strings"
)

// Relative file path used to store secrets in the config sent to Halyard
const SecretRelativeFilenames = "secrets"

type validationResponse []struct {
	Message  string `json:"message,omitempty"`
	Severity string `json:"severity,omitempty"`
	Location string `json:"location,omitempty"`
}

func (s *Service) Validate(ctx context.Context, spinsvc interfaces.SpinnakerService, failFast bool, logger logr.Logger) error {
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

func (s *Service) buildValidationRequest(ctx context.Context, spinsvc interfaces.SpinnakerService, failFast bool) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add config
	cfg := spinsvc.GetSpec().SpinnakerConfig

	// Sanitize secrets before validating
	// This will also serve as a secret validation step
	sanitizedCfg, err := sanitizeSecrets(ctx, cfg.Config)
	if err != nil {
		return nil, err
	}

	if err := s.addObjectToRequest(writer, "config", sanitizedCfg); err != nil {
		return nil, err
	}
	// Add required files
	for k := range cfg.Files {
		if err := s.addPart(writer, k, cfg.GetFileContent(k)); err != nil {
			return nil, err
		}
	}
	// Add cached secret files
	secCtx, err := secrets.FromContextWithError(ctx)
	if err != nil {
		return nil, err
	}

	for _, f := range secCtx.FileCache {
		// Get the key
		k := fmt.Sprintf("%s__%s", SecretRelativeFilenames, path.Base(f))
		b, err := ioutil.ReadFile(f)
		if err != nil {
			return nil, err
		}
		if err := s.addPart(writer, k, b); err != nil {
			return nil, err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v1/validation/config?failFast=%t&skipValidators=%s", s.url, failFast, strings.Join(getValidationsToSkip(spinsvc.GetSpec().Validation), ","))
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return req, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

// sanitizeSecrets inspects the given interface and decrypt any secret
// Tokens are decrypted, files are added to the secret context and replaced
// with SecretRelativeFilenames/tempfilename
func sanitizeSecrets(ctx context.Context, object interface{}) (interface{}, error) {
	h := func(val string) (string, error) {
		if secrets.ShouldDecryptToValidate(val) {
			s, f, err := secrets.Decode(ctx, val)
			if err == nil && f {
				s = path.Join(SecretRelativeFilenames, path.Base(s))
			}
			return s, err
		}
		return val, nil
	}
	return inspect.InspectStrings(object, h)
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
