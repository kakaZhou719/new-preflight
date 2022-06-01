package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

)

// PortSlice port slice
type PortSlice []*Port

func (s PortSlice) Value() (driver.Value, error) {
	result, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return driver.Value(string(result)), nil
}

func (s *PortSlice) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, &s)
	case string:
		return json.Unmarshal([]byte(v), &s)
	default:
		return fmt.Errorf("cannot sql.Scanner.Scan() PortSlice from: %#v", v)
	}
	return nil
}
