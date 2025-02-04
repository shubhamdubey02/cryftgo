// Copyright (C) 2019-2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package tmpnet

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cast"

	"github.com/shubhamdubey02/cryftgo/utils/perms"
)

// Defines a mapping of flag keys to values intended to be supplied to
// an invocation of an CryftGo node.
type FlagsMap map[string]interface{}

// Utility function simplifying construction of a FlagsMap from a file.
func ReadFlagsMap(path string, description string) (FlagsMap, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", description, err)
	}
	flagsMap := FlagsMap{}
	if err := json.Unmarshal(bytes, &flagsMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %w", description, err)
	}
	return flagsMap, nil
}

// SetDefaults ensures the effectiveness of flag overrides by only
// setting values supplied in the defaults map that are not already
// explicitly set.
func (f FlagsMap) SetDefaults(defaults FlagsMap) {
	for key, value := range defaults {
		if _, ok := f[key]; !ok {
			f[key] = value
		}
	}
}

// GetStringVal simplifies retrieving a map value as a string.
func (f FlagsMap) GetStringVal(key string) (string, error) {
	rawVal, ok := f[key]
	if !ok {
		return "", nil
	}

	val, err := cast.ToStringE(rawVal)
	if err != nil {
		return "", fmt.Errorf("failed to cast value for %q: %w", key, err)
	}
	return val, nil
}

// Write simplifies writing a FlagsMap to the provided path. The
// description is used in error messages.
func (f FlagsMap) Write(path string, description string) error {
	bytes, err := DefaultJSONMarshal(f)
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", description, err)
	}
	if err := os.WriteFile(path, bytes, perms.ReadWrite); err != nil {
		return fmt.Errorf("failed to write %s: %w", description, err)
	}
	return nil
}
