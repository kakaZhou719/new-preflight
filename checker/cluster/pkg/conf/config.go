package conf

import (
	"fmt"
	"preflight/pkg/sshcmd/sshutil"
)

var ScriptDir string

var SshConfig = sshutil.SSH{}

// OS contains the OS name and the kernel version
type OS struct {
	OSName        string `json:"osName,omitempty" yaml:"osName,omitempty"`
	OSVersion     string `json:"osVersion,omitempty" yaml:"osVersion,omitempty"`
	KernelVersion string `json:"kernelVersion,omitempty" yaml:"kernelVersion,omitempty"`
	Arch          string `json:"arch,omitempty" yaml:"arch,omitempty"`
}

func (sro *OS) GetStr() string {
	return fmt.Sprintf("{Arch:%s, OS:%s, Version:%s, Kernel:%s}", sro.Arch, sro.OSName, sro.OSVersion, sro.KernelVersion)
}

var SupportedOS = []OS{
	{OSName: "CentOS", OSVersion: "7.7", KernelVersion: "3.10.0", Arch: "amd64"},
	{OSName: "CentOS", OSVersion: "7.8", KernelVersion: "3.10.0", Arch: "amd64"},
	{OSName: "CentOS", OSVersion: "8.2", KernelVersion: "4.18", Arch: "amd64"},
}

func GetSupportedOSStr() []string {
	arr := make([]string, 0)
	for _, sro := range SupportedOS {
		arr = append(arr, fmt.Sprintf("{Arch:%s, OS:%s, Version:%s.*, Kernel:%s.*}", sro.Arch, sro.OSName, sro.OSVersion, sro.KernelVersion))
	}
	return arr
}

type HardwareResource struct {
	CpuMinimum        int32
	MemMinimum        int32
	SystemDiskMinimum int32
}

var HardwareResourceRequired = HardwareResource{
	CpuMinimum:        4,
	MemMinimum:        8,
	SystemDiskMinimum: 100,
}
