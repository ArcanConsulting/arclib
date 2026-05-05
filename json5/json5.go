package json5

import "encoding/json"

// Unmarshal parses JSON5 data and stores the result in the value pointed to by v.
// It works by normalizing the JSON5 input to valid JSON, then delegating to
// encoding/json.Unmarshal.
func Unmarshal(data []byte, v interface{}) error {
	normalized := normalize(data)
	return json.Unmarshal(normalized, v)
}

// UnmarshalToMap parses JSON5 data into a map[string]interface{}.
// This is a convenience function for configuration file parsing where
// the structure is not known at compile time.
func UnmarshalToMap(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}
