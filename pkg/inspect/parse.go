package inspect

import (
	"encoding/json"
)

func Convert(i1 interface{}, i2 interface{}) error {
	b, err := json.Marshal(i1)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, i2)
}

// Dispatch dispatches keys in the input settings onto the 3 objects passed in the order defined
func Dispatch(settings map[string]interface{}, obj ...interface{}) error {
	b, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	for i := range obj {
		if err := json.Unmarshal(b, obj[i]); err != nil {
			return err
		}
	}
	return nil
}
