package xtoml

import (
	"fmt"

	"github.com/pelletier/go-toml/v2"
)

func UnmarshalMap(data map[string]any, v any) error {
	rawData, err := toml.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal config data: %w", err)
	}

	if err = toml.Unmarshal(rawData, v); err != nil {
		return fmt.Errorf("unmarshal config data: %w", err)
	}

	return nil
}
