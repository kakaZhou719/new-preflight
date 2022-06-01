package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// StringMap string map
type StringMap map[string]string

func (sm StringMap) Value() (driver.Value, error) {
	result, err := json.Marshal(sm)
	if err != nil {
		return "", err
	}
	return driver.Value(string(result)), nil
}

func (sm *StringMap) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, &sm)
	case string:
		return json.Unmarshal([]byte(v), &sm)
	default:
		return fmt.Errorf("cannot sql.Scanner.Scan() StringMap from: %#v", v)
	}
	return nil
}
