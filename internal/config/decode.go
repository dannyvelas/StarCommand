package config

import (
	"encoding/json"
	"fmt"
)

func decode(src, dest any) error {
	bytes, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("error marshalling map: %v", err)
	}

	if err := json.Unmarshal(bytes, dest); err != nil {
		return fmt.Errorf("error unmarshalling map into target: %v", err)
	}
	return nil
}
