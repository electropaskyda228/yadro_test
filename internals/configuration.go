package internals

import (
	"encoding/json"
	"os"
)

func GetConfiguration(path string) (*Race, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var race Race
	err = json.Unmarshal(data, &race)
	if err != nil {
		return nil, err
	}

	return &race, nil
}
