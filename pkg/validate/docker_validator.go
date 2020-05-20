package validate

import (
	"context"
	"errors"
	"fmt"
	tools "github.com/armory/go-yaml-tools/pkg/secrets"
	"github.com/armory/spinnaker-operator/pkg/apis/spinnaker/interfaces"
	"github.com/armory/spinnaker-operator/pkg/inspect"
	"github.com/armory/spinnaker-operator/pkg/secrets"
	"github.com/armory/spinnaker-operator/pkg/util"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"unicode"
)

//go:generate mockgen -destination=docker_validator_mocks_test.go -package validate -source docker_validator.go

const (
	dockerRegistryAccountType        = "docker"
	dockerRegistryAccountsEnabledKey = "providers.dockerRegistry.enabled"
	dockerRegistryAccountsKey        = "providers.dockerRegistry.accounts"
	namePattern                      = "^[a-z0-9]+([-a-z0-9]*[a-z0-9])?$"
)

type dockerRegistryAccount struct {
	Name                    string                 `json:"name,omitempty"`
	RequiredGroupMembership []string               `json:"requiredGroupMembership,omitempty"`
	ProviderVersion         string                 `json:"providerVersion,omitempty"`
	Permissions             map[string]interface{} `json:"permissions,omitempty"`
	Address                 string                 `json:"address,omitempty"`
	Username                string                 `json:"username,omitempty"`
	Password                string                 `json:"password,omitempty"`
	PasswordCommand         string                 `json:"passwordCommand,omitempty"`
	Email                   string                 `json:"email,omitempty"`
	CacheIntervalSeconds    int64                  `json:"cacheIntervalSeconds,omitempty"`
	ClientTimeoutMillis     int64                  `json:"clientTimeoutMillis,omitempty"`
	CacheThreads            int32                  `json:"cacheThreads,omitempty"`
	PaginateSize            int32                  `json:"paginateSize,omitempty"`
	SortTagsByDate          bool                   `json:"sortTagsByDate,omitempty"`
	TrackDigests            bool                   `json:"trackDigests,omitempty"`
	InsecureRegistry        bool                   `json:"insecureRegistry,omitempty"`
	Repositories            []string               `json:"repositories,omitempty"`
	PasswordFile            string                 `json:"passwordFile,omitempty"`
	DockerconfigFile        string                 `json:"dockerconfigFile,omitempty"`
}

func (d *dockerRegistryAccount) GetAddress() string {
	if strings.HasPrefix(d.Address, "https://") || strings.HasPrefix(d.Address, "http://") {
		return d.Address
	} else {
		return fmt.Sprintf("https://%s", d.Address)
	}
}

type dockerRegistryValidator struct{}

func (d *dockerRegistryValidator) Validate(spinSvc interfaces.SpinnakerService, options Options) ValidationResult {

	accountEnabled, err := spinSvc.GetSpinnakerConfig().GetHalConfigPropBool(dockerRegistryAccountsEnabledKey, false)
	if err != nil {
		return ValidationResult{}
	}

	if !d.validationEnabled(spinSvc.GetSpinnakerValidation()) || !accountEnabled {
		return ValidationResult{}
	}

	dockerRegistries, err := spinSvc.GetSpinnakerConfig().GetHalConfigObjectArray(options.Ctx, dockerRegistryAccountsKey)
	if err != nil {
		// Ignore, key or format don't match expectations
		return ValidationResult{}
	}

	for _, rm := range dockerRegistries {

		var registry dockerRegistryAccount
		if err := mapstructure.Decode(rm, &registry); err != nil {
			return NewResultFromError(err, true)
		}

		if ok, err := d.validateRegistry(registry, options.Ctx, spinSvc); !ok {
			return NewResultFromErrors(err, true)
		}
	}

	return ValidationResult{}
}

func (d *dockerRegistryValidator) validationEnabled(v *interfaces.SpinnakerValidation) bool {
	for n, s := range v.Providers {
		if strings.ToLower(n) == strings.ToLower(dockerRegistryAccountType) {
			return s.Enabled
		}
	}
	return v.GetValidationSettings().Enabled
}

func (d *dockerRegistryValidator) validateRegistry(registry dockerRegistryAccount, ctx context.Context, spinSvc interfaces.SpinnakerService) (bool, []error) {

	var errs []error
	if len(registry.Name) == 0 {
		err := errors.New("dockerRegistry account missing name")
		return false, append(errs, err)
	}

	if len(regexp.MustCompile(namePattern).FindStringSubmatch(registry.Name)) == 0 {
		err := fmt.Errorf("Account name must match pattern %s\nIt must start and end with a lower-case character or number, and only contain lower-case characters, numbers, or dashes", namePattern)
		return false, append(errs, err)
	}

	resolvedPassword := ""
	passwordProvided := len(registry.Password) != 0
	passwordCommandProvided := len(registry.PasswordCommand) != 0
	passwordFileProvided := len(registry.PasswordFile) != 0

	if passwordProvided && passwordFileProvided || passwordCommandProvided && passwordProvided || passwordCommandProvided && passwordFileProvided {
		err := errors.New("You have provided more than one of password, password command, or password file for your docker registry. You can specify at most one.")
		return false, append(errs, err)
	}

	if passwordProvided {
		password, err := inspect.GetObjectPropString(ctx, registry, "Password")
		if err == nil {
			resolvedPassword = password
		}
	} else if passwordFileProvided {
		pf, err := inspect.GetObjectPropString(ctx, registry, "PasswordFile")
		if err != nil {
			return false, append(errs, err)
		}
		password, err := d.loadPasswordFromFile(pf, ctx, spinSvc.GetSpinnakerConfig())

		if err == nil {
			resolvedPassword = password
		}
		if len(resolvedPassword) == 0 || err != nil {
			err := errors.New("The supplied password file is empty.")
			return false, append(errs, err)
		}
	} else if passwordCommandProvided {
		out, err := exec.Command("bash", "-c", registry.PasswordCommand).Output()

		if err != nil {
			err := fmt.Errorf("password command returned non 0 return code, stderr/stdout was: %s", err)
			return false, append(errs, err)
		}

		resolvedPassword = strings.Trim(string(out), "\n")
		if len(resolvedPassword) == 0 {
			err := fmt.Errorf("Resolved password was empty, missing dependencies for running password command?")
			return false, append(errs, err)
		}

	}

	if len(resolvedPassword) != 0 && len(registry.Username) == 0 {
		err := errors.New("You have supplied a password but no username.")
		return false, append(errs, err)
	} else if len(resolvedPassword) == 0 && len(registry.Username) != 0 {
		err := errors.New("You have a supplied a username but no password.")
		return false, append(errs, err)
	}

	service := dockerRegistryService{address: registry.GetAddress(), username: registry.Username, password: resolvedPassword, httpService: util.HttpService{}, ctx: ctx}

	if registry.Username != "" && registry.Password != "" {
		ok, err := service.GetBase()

		if err != nil {
			return false, append(errs, err)
		}

		if !ok {
			if len(resolvedPassword) != 0 {
				c := resolvedPassword[len(resolvedPassword)-1]
				if unicode.IsSpace(rune(c)) {
					err := errors.New("Your password file has a trailing newline; many text editors append a newline to files they open." + " If you think this is causing authentication issues, you can strip the newline with the command:\n\n" + " tr -d '\\n' < PASSWORD_FILE | tee PASSWORD_FILE")
					return false, append(errs, err)
				}
			}

			err := errors.New(fmt.Sprintf("Unable to establish a connection with docker registry %s with provided credentials", registry.GetAddress()))
			return false, append(errs, err)
		}
	}

	if len(registry.Repositories) != 0 {
		v := newDockerRepoValidator(ctx)
		repositoryErrors := v.repository(registry, &service)
		for _, err := range repositoryErrors {
			if err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			return false, errs
		}
	}

	return true, nil
}

type dockerRepositoryValidate struct {
	ctx                 context.Context
	repositoryValidator dockerRepositoryValidator
}

type dockerRepositoryValidator interface {
	repository(registry dockerRegistryAccount, service *dockerRegistryService) []error
	imageTags(repository string, service *dockerRegistryService) error
}

func newDockerRepoValidator(ctx context.Context) *dockerRepositoryValidate {
	v := dockerRepositoryValidate{ctx: ctx, repositoryValidator: &dockerRepositoryValidate{}}
	return &v
}

func (v *dockerRepositoryValidate) repository(registry dockerRegistryAccount, service *dockerRegistryService) []error {
	wg := sync.WaitGroup{}
	var errs []error
	for _, r := range registry.Repositories {
		wg.Add(1)

		go func(r string) {
			err := v.repositoryValidator.imageTags(r, service)
			errs = append(errs, err)
			wg.Done()
		}(r)
	}

	wg.Wait()

	return errs
}

func (d *dockerRepositoryValidate) imageTags(repository string, service *dockerRegistryService) error {
	tagCount, err := service.GetTagsCount(repository)

	if err != nil {
		return err
	}

	if tagCount == 0 {
		return fmt.Errorf("Repository %s contain any tags. Spinnaker will not be able to deploy any docker images, Push some images to your registry.", repository)
	}

	return nil
}

func (d *dockerRegistryValidator) loadPasswordFromFile(passwordFile string, ctx context.Context, spinCfg *interfaces.SpinnakerConfig) (string, error) {
	if tools.IsEncryptedSecret(passwordFile) {
		path, err := secrets.DecodeAsFile(ctx, passwordFile)
		if err != nil {
			return "", err
		}
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return "", err
		}
		return string(content), nil
	} else if filepath.IsAbs(passwordFile) {
		// if file path is absolute, it may already be a path decoded by secret engines
		content, err := ioutil.ReadFile(passwordFile)
		if err != nil {
			return "", err
		}
		return string(content), nil
	} else {
		// we're taking relative file paths as files defined inside spec.spinnakerConfig.files
		content := spinCfg.GetFileContent(passwordFile)
		return string(content), nil
	}
}
