package types

type ClusterInfoBrief struct {
	SshUser     string   `json:"sshUser" yaml:"sshUser"`
	SshPassword string   `json:"sshPassword" yaml:"sshPassword"`
	Hosts       []string `json:"hosts" yaml:"hosts"`
}

type ClusterInfoDetailed struct {
	ClusterInfo   ClusterScopeInfo `json:",inline" yaml:",inline"`
	InstanceInfos []InstanceInfo   `json:"instance_list" yaml:"instance_list"`
}

type ClusterScopeInfo struct {
}
