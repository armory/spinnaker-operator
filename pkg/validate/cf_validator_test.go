package validate

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_cloudFoundryValidator_Validate_Account_Name(t *testing.T) {

	// given
	spinsvc, err := getSpinnakerService()
	if !assert.Nil(t, err) {
		return
	}
	cfValidator := cloudFoundryValidator{}
	cfAccount := cloudFoundryAccount{}

	// when
	ok, errs := cfValidator.validateAccount (cfAccount, context.TODO(), spinsvc)

	// then
	assert.Equal(t, false, ok)
	assert.Contains(t, fmt.Sprintf("%v", errs), "missing account name")

}

func Test_cloudFoundryValidator_Validate_Account_Name_Pattern(t *testing.T) {

	// given
	spinsvc, err := getSpinnakerService()
	if !assert.Nil(t, err) {
		return
	}
	cfValidator := cloudFoundryValidator{}
	cfAccount := cloudFoundryAccount{Name: "abcCloudFoundry"}

	// when
	ok, errs := cfValidator.validateAccount (cfAccount, context.TODO(), spinsvc)

	// then
	assert.Equal(t, false, ok)
	assert.Contains(t, fmt.Sprintf("%v", errs), "Account name must match pattern ^[a-z0-9]+([-a-z0-9]*[a-z0-9])?$")

}

func Test_cloudFoundryValidator_Validate_Account_UsernameAndPassword(t *testing.T) {

	// given
	spinsvc, err := getSpinnakerService()
	if !assert.Nil(t, err) {
		return
	}
	cfValidator := cloudFoundryValidator{}
	cfAccount := cloudFoundryAccount{Name: "dev"}

	// when
	ok, errs := cfValidator.validateAccount (cfAccount, context.TODO(), spinsvc)

	// then
	assert.Equal(t, false, ok)
	assert.Contains(t, fmt.Sprintf("%v", errs), "You must provide a user and a password")

}

func Test_cloudFoundryValidator_Validate_Account_UsernameButNoPassword(t *testing.T) {

	// given
	spinsvc, err := getSpinnakerService()
	if !assert.Nil(t, err) {
		return
	}
	cfValidator := cloudFoundryValidator{}
	cfAccount := cloudFoundryAccount{Name: "dev", User: "admin"}

	// when
	ok, errs := cfValidator.validateAccount (cfAccount, context.TODO(), spinsvc)

	// then
	assert.Equal(t, false, ok)
	assert.Contains(t, fmt.Sprintf("%v", errs), "You must provide a user and a password")

}

func Test_cloudFoundryValidator_Validate_Account_API(t *testing.T) {

	// given
	spinsvc, err := getSpinnakerService()
	if !assert.Nil(t, err) {
		return
	}
	cfValidator := cloudFoundryValidator{}
	cfAccount := cloudFoundryAccount{Name: "dev", User: "admin", Password: "123password", Api: "invalidApi.com"}

	// when
	ok, errs := cfValidator.validateAccount (cfAccount, context.TODO(), spinsvc)

	// then
	assert.Equal(t, false, ok)
	assert.Contains(t, fmt.Sprintf("%v", errs), "API must match pattern")

}
