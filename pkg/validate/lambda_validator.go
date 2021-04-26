package validate

import (
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"os"
	"strconv"
)

const (
	lambdaAccountType        		= "lambda"
	lambdaClouddriverEnabledKey 	= "aws.features.lambda.enabled"
	AccessKeyId 					= "providers.aws.accessKeyId"
	SecretAccessKey 				= "providers.aws.secretAccessKey"
)

type lambdaValidator struct{}

func (d *lambdaValidator) Validate(spinSvc interfaces.SpinnakerService, options Options) ValidationResult {

	//AWS
	//Check if the AWS Provider is enabled
	awsAccountEnabled, err := spinSvc.GetSpinnakerConfig().GetHalConfigPropBool(awsAccountsEnabledKey, false)
	if err != nil {
		return ValidationResult{}
	}

	if !spinSvc.GetSpinnakerValidation().IsProviderValidationEnabled(awsAccountType) || !awsAccountEnabled {
		return ValidationResult{}
	}

	//Get the AccessKeyId
	awsAccessKey, err := spinSvc.GetSpinnakerConfig().GetHalConfigPropString(options.Ctx, AccessKeyId)
	if err != nil {
		return ValidationResult{Errors: []error{fmt.Errorf("AccessKeyId is missing")}}
	}

	//Get the SecretAccessKey
	awsSecretKey, err := spinSvc.GetSpinnakerConfig().GetHalConfigPropString(options.Ctx, SecretAccessKey)
	if err != nil {
		return ValidationResult{Errors: []error{fmt.Errorf("SecretAccessKey is missing")}}
	}

	//Lambda
	isEnabled, err := spinSvc.GetSpinnakerConfig().GetRawServiceConfigPropString("clouddriver",lambdaClouddriverEnabledKey)
	if err != nil {
		return ValidationResult{}
	}
	accountEnabled, err := strconv.ParseBool(isEnabled)

	if !spinSvc.GetSpinnakerValidation().IsProviderValidationEnabled(lambdaAccountType) || !accountEnabled {
		return ValidationResult{}
	}

	//Get Regions
	regions, err :=  spinSvc.GetSpinnakerConfig().GetHalConfigObjectArray(options.Ctx, "providers.aws.defaultRegions")
	if err != nil {
		return ValidationResult{Errors: []error{fmt.Errorf("default regions is required")}}
	}

	for _, remap := range regions {
		for _, r := range remap {
			if ok, err := d.validateAWSLambda(awsAccessKey, awsSecretKey, r.(string)); !ok {
				return NewResultFromErrors(err, true)
			}
		}
	}

	return ValidationResult{}
}

func (d *lambdaValidator) validateAWSLambda(accessKey string, secretKey string, region string ) (bool, []error) {

	os.Setenv("AWS_ACCESS_KEY_ID",     accessKey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", secretKey)

	conf := aws.Config{Region: aws.String(region)}
	sess := session.New(&conf)
	svc := lambda.New(sess)
	input := &lambda.ListFunctionsInput{}

	_, err := svc.ListFunctions(input)
	if err != nil {
    	return false, []error{fmt.Errorf(err.Error())}
	}
	fmt.Println("Lambda Validation Passed")
	return true, nil
}
