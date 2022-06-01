package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type NetWorkCard struct {
	// Name
	Name string `json:"name" yaml:"name"`

	// IP
	IP string `json:"ip" yaml:"ip"`

	// MAC
	MAC string `json:"mac" yaml:"mac"`
}

func (s NetWorkCard) Value() (driver.Value, error) {
	result, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return driver.Value(string(result)), nil
}

func (s *NetWorkCard) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, &s)
	case string:
		return json.Unmarshal([]byte(v), &s)
	default:
		return fmt.Errorf("cannot sql.Scanner.Scan() NetWorkCard from: %#v", v)
	}
	return nil
}

type NetWorkCardSlice []*NetWorkCard

func (s NetWorkCardSlice) Value() (driver.Value, error) {
	result, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return driver.Value(string(result)), nil
}

func (s *NetWorkCardSlice) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, &s)
	case string:
		return json.Unmarshal([]byte(v), &s)
	default:
		return fmt.Errorf("cannot sql.Scanner.Scan() NetWorkCardSlice from: %#v", v)
	}
	return nil
}
