package account

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
)

type BaseAccount struct{}

func (b *BaseAccount) BaseToSpinnakerSettings(a Account) map[string]interface{} {
	m := make(map[string]interface{})
	m["name"] = a.GetName()
	if a.GetSettings() != nil {
		for key, val := range *a.GetSettings() {
			m[key] = val
		}
	}
	return m
}

func (b *BaseAccount) GetHash() (string, error) {
	data, err := json.Marshal(b)
	if err != nil {
		return "", err
	}
	m := md5.Sum(data)
	return hex.EncodeToString(m[:]), nil
}
