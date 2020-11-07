package validate

import (
	"reflect"
	"testing"
)

func Test_awsLifecycleHookValidation_isValidDefaultResult(t *testing.T) {
	type args struct {
		defaultResult string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Validation should fail, since DefaultResult is invalid",
			args: args{
				defaultResult: "OTHER",
			},
			want: false,
		},
		{
			name: "Validation should pass, since DefaultResult is valid 'CONTINUE'",
			args: args{
				defaultResult: "CONTINUE",
			},
			want: true,
		},
		{
			name: "Validation should pass, since DefaultResult is valid 'ABANDON'",
			args: args{
				defaultResult: "ABANDON",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &awsLifecycleHookValidation{}
			if got := a.isValidDefaultResult(tt.args.defaultResult); got != tt.want {
				t.Errorf("isValidDefaultResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_awsLifecycleHookValidation_isValidHeartbeatTimeout(t *testing.T) {
	type args struct {
		timeout int32
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Validation should fail, since HeartbeatTimeout is invalid",
			args: args{
				timeout: 0,
			},
			want: false,
		},
		{
			name: "Validation should fail, since HeartbeatTimeout is over the limit",
			args: args{
				timeout: 7201,
			},
			want: false,
		},
		{
			name: "Validation should fail, since HeartbeatTimeout is under the limit",
			args: args{
				timeout: 29,
			},
			want: false,
		},
		{
			name: "Validation should pass, since HeartbeatTimeout is valid",
			args: args{
				timeout: 30,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &awsLifecycleHookValidation{}
			if got := a.isValidHeartbeatTimeout(tt.args.timeout); got != tt.want {
				t.Errorf("isValidHeartbeatTimeout() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_awsLifecycleHookValidation_isValidRoleArn(t *testing.T) {
	type args struct {
		arn string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Validation should fail, since RoleARN is invalid",
			args: args{
				arn: "arn:aws:fail::11111111:role/test-aws-operator-validation-topic-role",
			},
			want: false,
		},
		{
			name: "Validation should pass, since RoleARN is valid",
			args: args{
				arn: "arn:aws:iam::11111111:role/test-aws-operator-validation-topic-role",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &awsLifecycleHookValidation{}
			if got := a.isValidRoleArn(tt.args.arn); got != tt.want {
				t.Errorf("isValidRoleArn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_awsLifecycleHookValidation_isValidSnsArn(t *testing.T) {

	type args struct {
		arn string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Validation should fail, since NotificationTargetARN is invalid",
			args: args{
				arn: "arn:aws:fail:us-west-2:11111111:test-aws-operator-validation-topic",
			},
			want: false,
		},
		{
			name: "Validation should pass, since NotificationTargetARN is valid",
			args: args{
				arn: "arn:aws:sns:us-west-2:11111111:test-aws-operator-validation-topic",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &awsLifecycleHookValidation{}
			if got := a.isValidSnsArn(tt.args.arn); got != tt.want {
				t.Errorf("isValidSnsArn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_awsLifecycleHookValidation_validate(t *testing.T) {
	awsHook := AwsLifecycleHook{
		DefaultResult:         "CONTINUE",
		HeartbeatTimeout:      120,
		LifecycleTransition:   "autoscaling:EC2_INSTANCE_TERMINATING",
		NotificationTargetARN: "arn:aws:sns:us-west-2:11111111:test-aws-operator-validation-topic",
		RoleARN:               "arn:aws:iam::11111111:role/test-aws-operator-validation-topic-role",
	}
	type args struct {
		hook AwsLifecycleHook
	}
	tests := []struct {
		name string
		args args
		want []error
	}{
		{
			name: "Validation should pass",
			args: args{hook: awsHook},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &awsLifecycleHookValidation{}
			if got := a.validate(tt.args.hook); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("validate() = %v, want %v", got, tt.want)
			}
		})
	}
}
