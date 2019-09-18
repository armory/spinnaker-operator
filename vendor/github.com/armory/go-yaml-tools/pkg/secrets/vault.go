package secrets

import (
	"fmt"
	"github.com/hashicorp/vault/api"
	"io/ioutil"
	"os"
	"strings"
)

type VaultConfig struct {
	Enabled    bool   `json:"enabled" yaml:"enabled"`
	Url        string `json:"url" yaml:"url"`
	AuthMethod string `json:"authMethod" yaml:"authMethod"`
	Role       string `json:"role" yaml:"role"`
	Path       string `json:"path" yaml:"path"`
	Token      string
}

type VaultSecret struct {
	engine        string
	path          string
	key           string
	base64Encoded string
}

type VaultDecrypter struct {
	params map[string]string
}

func NewVaultDecrypter(params map[string]string) *VaultDecrypter {
	return &VaultDecrypter{params}
}

func (decrypter *VaultDecrypter) Decrypt() (string, error) {
	if (VaultConfig{}) == Registry.VaultConfig {
		return "", fmt.Errorf("error: vault secrets configuration not found")
	}
	vaultSecret, err := ParseVaultSecret(decrypter.params)
	if err != nil {
		return "", fmt.Errorf("error parsing vault secret syntax - %s", err)
	}

	if Registry.VaultConfig.Token == "" {
		token, err := decrypter.FetchVaultToken()
		if err != nil {
			return "", fmt.Errorf("error fetching vault token - %s", err)
		}
		Registry.VaultConfig.Token = token
	}

	secret, err := decrypter.FetchSecret(vaultSecret)
	if err != nil && strings.Contains(err.Error(), "403") {
		// get new token and retry in case our saved token is no longer valid
		return decrypter.RetryFetchSecret(vaultSecret)
	}
	return secret, err
}

func ValidateVaultConfig(vaultConfig VaultConfig) error {
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
		if token := os.Getenv("VAULT_TOKEN"); token == "" {
			return fmt.Errorf("VAULT_TOKEN environment variable not set")
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

func ParseVaultSecret(params map[string]string) (VaultSecret, error) {
	var vaultSecret VaultSecret

	engine, ok := params["e"]
	if !ok {
		return VaultSecret{}, fmt.Errorf("secret format error - 'e' for engine is required")
	}
	vaultSecret.engine = engine

	path, ok := params["p"]
	if !ok {
		path, ok = params["n"]
		if !ok {
			return VaultSecret{}, fmt.Errorf("secret format error - 'p' for path is required (replaces deprecated 'n' param)")
		}
	}
	vaultSecret.path = path

	key, ok := params["k"]
	if !ok {
		return VaultSecret{}, fmt.Errorf("secret format error - 'k' for key is required")
	}
	vaultSecret.key = key

	base64, ok := params["b"]
	if ok {
		vaultSecret.base64Encoded = base64
	}

	return vaultSecret, nil
}

func (decrypter *VaultDecrypter) FetchVaultToken() (string, error) {
	if Registry.VaultConfig.AuthMethod == "TOKEN" {
		return os.Getenv("VAULT_TOKEN"), nil
	} else if Registry.VaultConfig.AuthMethod == "KUBERNETES" {
		return decrypter.FetchServiceAccountToken()
	} else {
		return "", fmt.Errorf("unknown Vault auth method: %q", Registry.VaultConfig.AuthMethod)
	}
}

func (decrypter *VaultDecrypter) FetchServiceAccountToken() (string, error) {
	client, err := api.NewClient(&api.Config{
		Address: Registry.VaultConfig.Url,
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
		"role": Registry.VaultConfig.Role,
		"jwt":  token,
	}

	secret, err := client.Logical().Write("auth/" + Registry.VaultConfig.Path + "/login", data)
	if err != nil {
		return "", fmt.Errorf("error logging into vault using kubernetes auth: %s", err)
	}

	return secret.Auth.ClientToken, nil
}

func (decrypter *VaultDecrypter) FetchVaultClient(token string) (*api.Client, error) {
	client, err := api.NewClient(&api.Config{
		Address: Registry.VaultConfig.Url,
	})
	if err != nil {
		return nil, err
	}
	client.SetToken(token)
	return client, nil
}

func (decrypter *VaultDecrypter) FetchSecret(secret VaultSecret) (string, error) {
	client, err := decrypter.FetchVaultClient(Registry.VaultConfig.Token)
	if err != nil {
		return "", fmt.Errorf("error fetching vault client - %s", err)
	}

	secretMapping, err := client.Logical().Read(secret.engine + "/" + secret.path)
	if err != nil {
		if strings.Contains(err.Error(), "invalid character '<' looking for beginning of value") {
			// some connection errors aren't properly caught, and the vault client tries to parse <nil>
			return "", fmt.Errorf("error fetching secret from vault - check connection to the server: %s",
				Registry.VaultConfig.Url)
		}
		return "", fmt.Errorf("error fetching secret from vault: %s", err)
	}

	warnings := secretMapping.Warnings
	if warnings != nil {
		for i := range warnings {
			if strings.Contains(warnings[i], "Invalid path for a versioned K/V secrets engine") {
				// try again using K/V v2 path
				secretMapping, err = client.Logical().Read(secret.engine + "/data/" + secret.path)
				if err != nil {
					return "", fmt.Errorf("error fetching secret from vault: %s", err)
				} else if secretMapping == nil {
					return "", fmt.Errorf("couldn't find vault path %q under engine %q", secret.path, secret.engine)
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

		decrypted, ok := mapping[secret.key].(string)
		if !ok {
			return "", fmt.Errorf("error fetching key %q", secret.key)
		}
		return decrypted, nil
	}

	return "", nil
}

func (decrypter *VaultDecrypter) RetryFetchSecret(secret VaultSecret) (string, error) {
	token, err := decrypter.FetchVaultToken()
	if err != nil {
		return "", fmt.Errorf("error fetching vault token - %s", err)
	}
	Registry.VaultConfig.Token = token
	return decrypter.FetchSecret(secret)
}
