package validate

import (
	"context"
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_lambdaValidator_Validate_AWS_Provider_Is_Disabled(t *testing.T) {

	// given
	spinsvc := test.ManifestFileToSpinService("testdata/spinvc_lambda_aws_disabled.yml", t)

	awsLambdaValidator := lambdaValidator{}

	options := Options{Ctx: context.TODO()}

	// when
	result := awsLambdaValidator.Validate(spinsvc, options)

	// then
	assert.Nil (t, result.Errors)
}

func Test_lambdaValidator_Validate_AccessKeyId_Missing(t *testing.T) {

	// given
	spinsvc := test.ManifestFileToSpinService("testdata/spinvc_lambda_access_key.yml", t)

	awsLambdaValidator := lambdaValidator{}

	options := Options{Ctx: context.TODO()}

	// when
	result := awsLambdaValidator.Validate(spinsvc, options)

	// then
	assert.Contains(t, fmt.Sprintf("%v", result), "AccessKeyId is missing")
}

func Test_lambdaValidator_Validate_SecretKey_Missing(t *testing.T) {

	// given
	spinsvc := test.ManifestFileToSpinService("testdata/spinvc_lambda_secret_key.yml", t)

	awsLambdaValidator := lambdaValidator{}

	options := Options{Ctx: context.TODO()}

	// when
	result := awsLambdaValidator.Validate(spinsvc, options)

	// then
	assert.Contains(t, fmt.Sprintf("%v", result), "SecretAccessKey is missing")
}

func Test_lambdaValidator_Validate_Lambda_Is_Disabled(t *testing.T) {

	// given
	spinsvc := test.ManifestFileToSpinService("testdata/spinvc_lambda_disabled.yml", t)

	awsLambdaValidator := lambdaValidator{}

	options := Options{Ctx: context.TODO()}

	// when
	result := awsLambdaValidator.Validate(spinsvc, options)

	// then
	assert.Nil (t, result.Errors)
}

func Test_lambdaValidator_Validate_No_Regions(t *testing.T) {

	// given
	spinsvc := test.ManifestFileToSpinService("testdata/spinvc_lambda_no_regions.yml", t)

	awsLambdaValidator := lambdaValidator{}

	options := Options{Ctx: context.TODO()}

	// when
	result := awsLambdaValidator.Validate(spinsvc, options)

	// then
	assert.Contains(t, fmt.Sprintf("%v", result), "default regions is required")
}