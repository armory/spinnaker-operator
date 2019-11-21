package secrets

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/vault/api"
)

func RegisterVaultConfig(vaultConfig VaultConfig) error {
	if err := validateVaultConfig(vaultConfig); err != nil {
		return fmt.Errorf("vault configuration error - %s", err)
	}
	Engines["vault"] = func(ctx context.Context, isFile bool, params string) (Decrypter, error) {
		vd := &VaultDecrypter{isFile: isFile, vaultConfig: vaultConfig}
		if err := vd.parse(params); err != nil {
			return nil, err
		}
		return vd, nil
	}
	return nil
}

type VaultConfig struct {
	Enabled    bool   `json:"enabled" yaml:"enabled"`
	Url        string `json:"url" yaml:"url"`
	AuthMethod string `json:"authMethod" yaml:"authMethod"`
	Role       string `json:"role" yaml:"role"`
	Path       string `json:"path" yaml:"path"`
	Token      string
}

type VaultSecret struct {
}

type VaultDecrypter struct {
	engine        string
	path          string
	key           string
	base64Encoded string
	isFile        bool
	vaultConfig   VaultConfig
}

func (decrypter *VaultDecrypter) Decrypt() (string, error) {
	if decrypter.vaultConfig.Token == "" {
		err := decrypter.setToken()
		if err != nil {
			return "", err
		}
	}
	secret, err := decrypter.fetchSecret()
	if err != nil && strings.Contains(err.Error(), "403") {
		// get new token and retry in case our saved token is no longer valid
		err := decrypter.setToken()
		if err != nil {
			return "", err
		}
		secret, err = decrypter.fetchSecret()
	}
	if err != nil || !decrypter.isFile {
		return secret, err
	}
	return ToTempFile([]byte(secret))
}

func (v *VaultDecrypter) IsFile() bool {
	return v.isFile
}

func (v *VaultDecrypter) parse(params string) error {
	tokens := strings.Split(params, "!")
	for _, element := range tokens {
		kv := strings.Split(element, ":")
		if len(kv) == 2 {
			switch kv[0] {
			case "e":
				v.engine = kv[1]
			case "p", "n":
				v.path = kv[1]
			case "k":
				v.key = kv[1]
			case "b":
				v.base64Encoded = kv[1]
			}
		}
	}

	if v.engine == "" {
		return fmt.Errorf("secret format error - 'e' for engine is required")
	}
	if v.path == "" {
		return fmt.Errorf("secret format error - 'p' for path is required (replaces deprecated 'n' param)")
	}
	if v.key == "" {
		return fmt.Errorf("secret format error - 'k' for key is required")
	}
	return nil
}

func (decrypter *VaultDecrypter) setToken() error {
	var token string
	var err error

	if decrypter.vaultConfig.AuthMethod == "TOKEN" {
		token = os.Getenv("VAULT_TOKEN")
		if token == "" {
			return fmt.Errorf("VAULT_TOKEN environment variable not set")
		}
	} else if decrypter.vaultConfig.AuthMethod == "KUBERNETES" {
		token, err = decrypter.fetchServiceAccountToken()
	} else {
		err = fmt.Errorf("unknown Vault auth method: %q", decrypter.vaultConfig.AuthMethod)
	}

	if err != nil {
		return fmt.Errorf("error fetching vault token - %s", err)
	}
	decrypter.vaultConfig.Token = token
	return nil
}

func validateVaultConfig(vaultConfig VaultConfig) error {
	if (VaultConfig{}) == vaultConfig {
		return fmt.Errorf("vault secrets not configured in service profile yaml")
	}
	if vaultConfig.Enabled == false {
		return fmt.Errorf("vault secrets disabled")
	}
	if vaultConfig.Url == "" {
		return fmt.Errorf("vault url required")
	}
	if vaultConfig.AuthMethod == "" {
		return fmt.Errorf("auth method required")
	}

	if vaultConfig.AuthMethod == "TOKEN" {
		if vaultConfig.Token == "" {
			envToken := os.Getenv("VAULT_TOKEN")
			if envToken == "" {
				return fmt.Errorf("VAULT_TOKEN environment variable not set")
			}
		}
	} else if vaultConfig.AuthMethod == "KUBERNETES" {
		if vaultConfig.Path == "" || vaultConfig.Role == "" {
			return fmt.Errorf("path and role both required for Kubernetes auth method")
		}
	} else {
		return fmt.Errorf("unknown Vault secrets auth method: %q", vaultConfig.AuthMethod)
	}

	return nil
}

func (decrypter *VaultDecrypter) fetchServiceAccountToken() (string, error) {
	client, err := api.NewClient(&api.Config{
		Address: decrypter.vaultConfig.Url,
	})
	if err != nil {
		return "", fmt.Errorf("error fetching vault client: %s", err)
	}

	tokenFile, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		return "", fmt.Errorf("error reading service account token: %s", err)
	}
	token := string(tokenFile)
	data := map[string]interface{}{
		"role": decrypter.vaultConfig.Role,
		"jwt":  token,
	}

	secret, err := client.Logical().Write("auth/"+decrypter.vaultConfig.Path+"/login", data)
	if err != nil {
		return "", fmt.Errorf("error logging into vault using kubernetes auth: %s", err)
	}

	return secret.Auth.ClientToken, nil
}

func (decrypter *VaultDecrypter) FetchVaultClient(token string) (*api.Client, error) {
	client, err := api.NewClient(&api.Config{
		Address: decrypter.vaultConfig.Url,
	})
	if err != nil {
		return nil, err
	}
	client.SetToken(token)
	return client, nil
}

func (decrypter *VaultDecrypter) fetchSecret() (string, error) {
	client, err := decrypter.FetchVaultClient(decrypter.vaultConfig.Token)
	if err != nil {
		return "", fmt.Errorf("error fetching vault client - %s", err)
	}

	secretMapping, err := client.Logical().Read(decrypter.engine + "/" + decrypter.path)
	if err != nil {
		if strings.Contains(err.Error(), "invalid character '<' looking for beginning of value") {
			// some connection errors aren't properly caught, and the vault client tries to parse <nil>
			return "", fmt.Errorf("error fetching secret from vault - check connection to the server: %s",
				decrypter.vaultConfig.Url)
		}
		return "", fmt.Errorf("error fetching secret from vault: %s", err)
	}

	warnings := secretMapping.Warnings
	if warnings != nil {
		for i := range warnings {
			if strings.Contains(warnings[i], "Invalid path for a versioned K/V secrets engine") {
				// try again using K/V v2 path
				secretMapping, err = client.Logical().Read(decrypter.engine + "/data/" + decrypter.path)
				if err != nil {
					return "", fmt.Errorf("error fetching secret from vault: %s", err)
				} else if secretMapping == nil {
					return "", fmt.Errorf("couldn't find vault path %q under engine %q", decrypter.path, decrypter.engine)
				}
				break
			}
		}
	}

	if secretMapping != nil {
		mapping := secretMapping.Data
		if data, ok := mapping["data"]; ok { // one more nesting of "data" if using K/V v2
			if submap, ok := data.(map[string]interface{}); ok {
				mapping = submap
			}
		}

		decrypted, ok := mapping[decrypter.key].(string)
		if !ok {
			return "", fmt.Errorf("error fetching key %q", decrypter.key)
		}
		return decrypted, nil
	}

	return "", nil
}
