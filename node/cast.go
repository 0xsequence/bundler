package node

import (
	"encoding/json"
	"fmt"
)

// TODO: fix this before release
func UglyCast[T any](untyped any) (T, error) {
	var typed T

	data, err := json.Marshal(untyped)
	if err != nil {
		return typed, fmt.Errorf("unable to marshal value: %w", err)
	}

	err = json.Unmarshal(data, &typed)
	if err != nil {
		return typed, fmt.Errorf("unable to unmarshal value: %w", err)
	}

	return typed, nil
}
