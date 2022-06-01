package types


type TimeSyncStatus struct {
	Ntpd    string `json:"ntpd" yaml:"ntpd"`
	Chronyd string `json:"chronyd" yaml:"chronyd"`
}

type InstanceInfoExtended struct {
	InstanceInfo      InstanceInfo `json:"instanceInfo" yaml:"instanceInfo"`
	TimeSyncStatus    TimeSyncStatus     `json:"timeSyncStatus" yaml:"timeSyncStatus"`
	NetworkDevicesStr string             `json:"networkDevicesStr" yaml:"networkDevicesStr"`
	BlockDevicesStr   string             `json:"blockDevicesStr" yaml:"blockDevicesStr"`
}


// InstanceInfo instance info
type InstanceInfo struct {
	// host name
	HostName string `json:"hostName,omitempty" yaml:"hostName,omitempty"`

	// identifier
	Identifier string `json:"identifier" yaml:"identifier" validate:"required"`

	// OS
	OS string `json:"os" yaml:"os"`

	// OS Version
	OSVersion string `json:"osVersion,omitempty" yaml:"osVersion,omitempty"`

	// Arch
	Arch string `json:"arch,omitempty" yaml:"arch,omitempty"`

	// Kernel
	Kernel string `json:"kernel" yaml:"kernel"`

	// MACAddress
	MACAddress string `json:"macAddress" yaml:"macAddress"`

	// cpu
	CPU int32 `json:"cpu" yaml:"cpu" validate:"required"`

	// memory
	Memory int32 `json:"memory" yaml:"memory" validate:"required"`

	// GPU related resources.
	GPU *GPU `json:"gpu,omitempty" yaml:"gpu,omitempty"`

	// ports
	ExposePorts PortSlice `json:"ports,omitempty" yaml:"ports,omitempty" gorm:"type:text"`

	// system disk
	SystemDisk DiskSlice `json:"systemDisk,omitempty" yaml:"systemDisk,omitempty" gorm:"type:varchar(1000)" validate:"required"`

	// data disk
	DataDisk DiskSlice `json:"dataDisk,omitempty" yaml:"dataDisk,omitempty" gorm:"type:varchar(1000)"`

	// private IP
	PrivateIP string `json:"privateIP" yaml:"privateIP" validate:"ip,required" validate:"required,ip"`

	// public IP
	PublicIP string `json:"publicIP,omitempty" yaml:"publicIP,omitempty" validate:"omitempty,ip"`

	// internet bandwidth
	InternetBandwidth int32 `json:"internetBandwidth,omitempty" yaml:"internetBandwidth,omitempty"`

	// NetworkCards
	NetworkCards NetWorkCardSlice `json:"networkCards" yaml:"networkCards" gorm:"type:text"`

	// image ID
	ImageID string `json:"imageID,omitempty" yaml:"imageID,omitempty"`

	// instance type
	InstanceType string `json:"instanceType,omitempty" yaml:"instanceType,omitempty"`

	// root password
	RootPassword string `json:"rootPassword" yaml:"rootPassword"`

	// annotations
	Annotations StringMap `json:"annotations,omitempty" yaml:"annotations,omitempty" gorm:"type:text"`
}

type Port struct {
	// The target IP address range.
	// Only IPv4 is supported.
	// The default value is 0.0.0.0/0(which means no restriction will be applied).
	//
	CidrIP string `json:"cidrIP" yaml:"cidrIP" validate:"cidrv4"`

	// The range of port numbers. The format should be: start/end. E.g. 53/65000.
	// And the valid value port range is (1,65535).
	// NOTE: when protocol equals all, the portRange will be always set to "-1/-1".
	//
	PortRange string `json:"portRange" yaml:"portRange"`

	// protocol
	Protocol Protocol `json:"protocol" yaml:"protocol" validate:"oneof=all tcp udp"`

	// The authorization policy.
	// It means to drop all requests from cidrIP access portRange when unallowed is true.
	//
	Unallowed bool `json:"unallowed" yaml:"unallowed"`
}

// Protocol The protocol type of port listening.
// swagger:model Protocol
type Protocol string

const (
	// ProtocolAll captures enum value "all"
	ProtocolAll Protocol = "all"

	// ProtocolTCP captures enum value "tcp"
	ProtocolTCP Protocol = "tcp"

	// ProtocolUDP captures enum value "udp"
	ProtocolUDP Protocol = "udp"
)

type PublicIP struct {
	// TODO: make it possible to config with unit
	// The unit is Mbps for now.
	//
	Bandwidth int32 `json:"bandwidth" yaml:"bandwidth"`

	// required
	Required int32 `json:"required" yaml:"required" validate:"required"`
}

type GPU struct {
	// Driver defines the GPU driver related information for installing GPU driver automatically.
	// If not defined, the driver will not be installed automatically.
	//
	Driver *GPUDriver `json:"driver,omitempty" yaml:"driver,omitempty" validate:"omitempty"`

	// required
	// Minimum: 1
	Required int32 `json:"required" yaml:"required" validate:"required,min=1"`

	// The spec type of GPU instance you expected such as "NVIDIA V100",etc.
	// TODO(starnop): use enum type
	//
	SpecType string `json:"specType" yaml:"specType" validate:"required"`
}

// GPUDriver g p u driver
// swagger:model GPUDriver
type GPUDriver struct {
	// CUDAVersion is the version of CUDA version.
	// The default version is 10.1.168.
	//
	CUDAVersion string `json:"CUDAVersion" yaml:"CUDAVersion" validate:"required"`

	// CuDNNVersion is the version of CUDA Deep Neural Network library.
	// The default value is 7.6.4.
	//
	CuDNNVersion string `json:"cuDNNVersion" yaml:"cuDNNVersion" validate:"required"`

	// DriverVersion is the GPU driver version.
	// The default value is 418.87.01.
	//
	DriverVersion string `json:"driverVersion" yaml:"driverVersion" validate:"required"`
}
