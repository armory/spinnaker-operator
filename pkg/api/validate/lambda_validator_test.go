package validate

import (
	"context"
	"fmt"
	"testing"

	"github.com/armory/spinnaker-operator/pkg/api/test"
	"github.com/stretchr/testify/assert"
)

func Test_lambdaValidator_Validate_AWS_Provider_Is_Disabled(t *testing.T) {

	// given
	spinsvc := test.ManifestFileToSpinService("testdata/spinvc_lambda_aws_disabled.yml", t)

	awsLambdaValidator := lambdaValidator{}

	options := Options{Ctx: context.TODO()}

	// when
	result := awsLambdaValidator.Validate(spinsvc, options)

	// then
	assert.Nil(t, result.Errors)
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
	assert.Nil(t, result.Errors)
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

func Test_lambdaValidator_Validate_Account_Disabled(t *testing.T) {

	// given
	spinsvc := test.ManifestFileToSpinService("testdata/spinvc_lambda_account_disabled.yml", t)

	awsLambdaValidator := lambdaValidator{}

	options := Options{Ctx: context.TODO()}

	// when
	result := awsLambdaValidator.Validate(spinsvc, options)

	// then
	assert.Nil(t, result.Errors)
}

func Test_lambdaValidator_Validate_AccountId_Missing(t *testing.T) {

	// given
	spinsvc := test.ManifestFileToSpinService("testdata/spinvc_lambda_account_id.yml", t)

	awsLambdaValidator := lambdaValidator{}

	options := Options{Ctx: context.TODO()}

	// when
	result := awsLambdaValidator.Validate(spinsvc, options)

	// then
	assert.Contains(t, fmt.Sprintf("%v", result), "aws accounts accountId is required")
}

func Test_lambdaValidator_Validate_Assume_Role_Missing(t *testing.T) {

	// given
	spinsvc := test.ManifestFileToSpinService("testdata/spinvc_lambda_assume_role.yml", t)

	awsLambdaValidator := lambdaValidator{}

	options := Options{Ctx: context.TODO()}

	// when
	result := awsLambdaValidator.Validate(spinsvc, options)

	// then
	assert.Contains(t, fmt.Sprintf("%v", result), "aws accounts assumeRole is required")
}

func Test_lambdaValidator_Validate_AAccount_Name_Missing(t *testing.T) {

	// given
	spinsvc := test.ManifestFileToSpinService("testdata/spinvc_lambda_account_name.yml", t)

	awsLambdaValidator := lambdaValidator{}

	options := Options{Ctx: context.TODO()}

	// when
	result := awsLambdaValidator.Validate(spinsvc, options)

	// then
	assert.Contains(t, fmt.Sprintf("%v", result), "aws accounts name is required")
}
