package validate

import (
	"fmt"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/mitchellh/mapstructure"
	"os"
	"strconv"
)

const (
	lambdaAccountType               = "lambda"
	lambdaClouddriverEnabledKey     = "aws.features.lambda.enabled"
	AccessKeyId                     = "providers.aws.accessKeyId"
	SecretAccessKey                 = "providers.aws.secretAccessKey"
	lambdaAccountsKey               = "aws.accounts"
)

type lambdaAccount struct {
	Name                    string                 `json:"name,omitempty"`
	LambdaEnabled           bool                   `json:"lambdaEnabled,omitempty"`
	AccountId               string                 `json:"accountId,omitempty"`
	AssumeRole              string                 `json:"assumeRole,omitempty"`
}

type lambdaValidator struct{}

func (d *lambdaValidator) Validate(spinSvc interfaces.SpinnakerService, options Options) ValidationResult {

	fmt.Println("Lambda Validator Start...")

	//Lambda
	isEnabled, err := spinSvc.GetSpinnakerConfig().GetRawServiceConfigPropString("clouddriver",lambdaClouddriverEnabledKey)
	if err != nil {
		return ValidationResult{}
	}
	accountEnabled, err := strconv.ParseBool(isEnabled)

	if !spinSvc.GetSpinnakerValidation().IsProviderValidationEnabled(lambdaAccountType) || !accountEnabled {
		return ValidationResult{}
	}

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
		//return ValidationResult{Errors: []error{fmt.Errorf("AccessKeyId is missing")}}
	}

	//Get the SecretAccessKey
	awsSecretKey, err := spinSvc.GetSpinnakerConfig().GetHalConfigPropString(options.Ctx, SecretAccessKey)
	if err != nil {
		//return ValidationResult{Errors: []error{fmt.Errorf("SecretAccessKey is missing")}}
	}

	//Get the Accounts
	lambdaAccounts, err := spinSvc.GetSpinnakerConfig().GetServiceConfigObjectArray("clouddriver", lambdaAccountsKey)
	if err != nil {
		// Ignore, key or format don't match expectations
		return ValidationResult{}
	}

	//Get Regions
	regions, err :=  spinSvc.GetSpinnakerConfig().GetHalConfigObjectArray(options.Ctx, "providers.aws.defaultRegions")
	if err != nil {
		return ValidationResult{Errors: []error{fmt.Errorf("default regions is required")}}
	}

	for _, rm := range lambdaAccounts {

		var lambdaAcc lambdaAccount
		if err := mapstructure.Decode(rm, &lambdaAcc); err != nil {
			return NewResultFromError(err, true)
		}

		if lambdaAcc.LambdaEnabled {
			if ok, err := d.validateAWSLambda(awsAccessKey, awsSecretKey, regions, lambdaAcc); !ok {
				return NewResultFromErrors(err, true)
			}
		}
	}

	return ValidationResult{}
}

func (d *lambdaValidator) validateAWSLambda(accessKey string, secretKey string, regions []map[string]interface{}, account lambdaAccount) (bool, []error) {

	if len(account.AccountId) <= 0{
		return false, []error{fmt.Errorf("aws accounts accountId is required")}
	}

	if len(account.AssumeRole) <= 0{
		return false, []error{fmt.Errorf("aws accounts assumeRole is required")}
	}

	if len(account.Name) <= 0{
		return false, []error{fmt.Errorf("aws accounts name is required")}
	}

	if len(accessKey) > 0 && len(secretKey) > 0 {
		os.Setenv("AWS_ACCESS_KEY_ID",     accessKey)
		os.Setenv("AWS_SECRET_ACCESS_KEY", secretKey)
	}

	for _, regionMap := range regions {
		for _, region := range regionMap {
			//conf := aws.Config{Region: aws.String(region.(string))}
			//sess := session.New(&conf)
			sess, err := session.NewSession(&aws.Config{
				Region: aws.String(region.(string)),
			})

			if err != nil {
				return false, []error{fmt.Errorf("NewSession Error: %s", err)}
			}

			// Create the credentials from AssumeRoleProvider to assume the role
			awsARN := "arn:aws:iam::"+account.AccountId+":"+account.AssumeRole
			cred := stscreds.NewCredentials(sess, awsARN)

			svc := lambda.New(sess, &aws.Config{Credentials: cred})
			//input := &lambda.ListFunctionsInput{}

			result, err := svc.ListFunctions(nil)
				if err != nil {
					if err, ok := err.(awserr.Error); ok {
						switch err.Code() {
						case "AccessDenied":
							return false, []error{fmt.Errorf("AccessDenied permission denied")}
						default:
							return false, []error{fmt.Errorf(err.Error())}
						}
					}
					return false, []error{fmt.Errorf(err.Error())}
				}
			for _, f := range result.Functions {
				fmt.Println("Name:        " + aws.StringValue(f.FunctionName))
				fmt.Println("Description: " + aws.StringValue(f.Description))
				fmt.Println("")
			}
			return true, nil
		}
	}
	return true, nil
}
