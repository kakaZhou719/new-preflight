package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Disk struct {
	// Name
	Name string `json:"name" yaml:"name"`

	// required
	// Minimum: 1
	// TODO: Deprecated
	Required int32 `json:"required,omitempty" yaml:"required,omitempty"`

	// Capacity the total storage capacity.
	Capacity int32 `json:"capacity" yaml:"capacity,omitempty"`

	// Remain the remain storage capacity.
	Remain int32 `json:"remain,omitempty" yaml:"remain,omitempty"`

	// FSType the file system type.
	FSType string `json:"fsType" yaml:"fsType"`

	// MountPoint
	MountPoint string `json:"mountPoint" yaml:"mountPoint"`

	// Type the disk type.
	Type string `json:"type" yaml:"type"`
}

func (s Disk) Value() (driver.Value, error) {
	result, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return driver.Value(string(result)), nil
}

func (s *Disk) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, &s)
	case string:
		return json.Unmarshal([]byte(v), &s)
	default:
		return fmt.Errorf("cannot sql.Scanner.Scan() Disk from: %#v", v)
	}
	return nil
}

// DiskSlice disk slice
type DiskSlice []*Disk

func (s DiskSlice) Value() (driver.Value, error) {
	result, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return driver.Value(string(result)), nil
}

func (s *DiskSlice) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, &s)
	case string:
		return json.Unmarshal([]byte(v), &s)
	default:
		return fmt.Errorf("cannot sql.Scanner.Scan() DiskSlice from: %#v", v)
	}
	return nil
}
