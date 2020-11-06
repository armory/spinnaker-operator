package validate

import (
	"fmt"
	"regexp"
)

const (
	snsRattern     = "^arn:aws:sns:[^:]+:[^:]+:[^:]+$"
	iamRolePattern = "^arn:aws:iam::[^:]+:[^:]+$"
)

var validLifecycleHookResults = []string{"ABANDON", "CONTINUE"}

type AwsLifecycleHook struct {
	DefaultResult         string `json:"defaultResult,omitempty"`
	HeartbeatTimeout      int32  `json:"heartbeatTimeout,omitempty"`
	LifecycleTransition   string `json:"lifecycleTransition,omitempty"`
	NotificationTargetARN string `json:"notificationTargetARN,omitempty"`
	RoleARN               string `json:"roleARN,omitempty"`
}

type awsLifecycleHookValidation struct{}

func (a *awsLifecycleHookValidation) validate(hook AwsLifecycleHook) []error {
	var errors []error

	if !a.isValidSnsArn(hook.NotificationTargetARN) {
		errors = append(errors, fmt.Errorf("invalid SNS notification ARN: %s", hook.NotificationTargetARN))
	}

	if !a.isValidRoleArn(hook.RoleARN) {
		errors = append(errors, fmt.Errorf("invalid IAM role ARN: %s", hook.RoleARN))
	}

	if !a.isValidDefaultResult(hook.DefaultResult) {
		errors = append(errors, fmt.Errorf("invalid lifecycle default result: %s", hook.DefaultResult))
	}

	if !a.isValidHeartbeatTimeout(hook.HeartbeatTimeout) {
		errors = append(errors, fmt.Errorf("lifecycle heartbeat timeout must be between 30 and 7200. Provided value was:  %v", hook.HeartbeatTimeout))
	}
	return errors
}

func (a *awsLifecycleHookValidation) isValidSnsArn(arn string) bool {
	if len(regexp.MustCompile(snsRattern).FindStringSubmatch(arn)) == 0 {
		return false
	}
	return true
}

func (a *awsLifecycleHookValidation) isValidRoleArn(arn string) bool {
	if len(regexp.MustCompile(iamRolePattern).FindStringSubmatch(arn)) == 0 {
		return false
	}
	return true
}

func (a *awsLifecycleHookValidation) isValidHeartbeatTimeout(timeout int32) bool {
	return timeout != 0 && timeout >= 30 && timeout <= 7200
}

func (a *awsLifecycleHookValidation) isValidDefaultResult(defaultResult string) bool {
	for _, item := range validLifecycleHookResults {
		if item == defaultResult {
			return true
		}
	}
	return false
}
