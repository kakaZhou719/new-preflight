package run

import (
	"fmt"
	"github.com/pkg/errors"
	"preflight/checker/cluster/api/types"
	"preflight/checker/cluster/pkg/conf"
	"preflight/pkg/sshcmd/sshutil"
	"strconv"
	"strings"
	"time"
)

func ValidateAll(brief *types.ClusterInfoBrief) (bool, error) {
	clusterInfoDetailed, instanceInfoExtends, err := ParseClusterInfo(brief)
	if err != nil {
		return false, fmt.Errorf("parse cluster info failed")
	}

	return ClusterValidator{
		Ssh:                      conf.SshConfig,
		SupportedOS:              conf.SupportedOS,
		HardwareResourceRequired: conf.HardwareResourceRequired,
	}.Validate(clusterInfoDetailed, instanceInfoExtends)
}

type ClusterValidator struct {
	Ssh                      sshutil.SSH
	SupportedOS              []conf.OS
	HardwareResourceRequired conf.HardwareResource
}

func (c ClusterValidator) Validate(detailed *types.ClusterInfoDetailed, instanceInfoExtends map[string]*types.InstanceInfoExtended) (bool, error) {
	// check clusterScopeInfo
	if err := c.validateHostName(detailed); err != nil {
		return false, errors.Errorf("validate hostname failed, error:%v", err)
	}

	// check time sync
	_, err := c.IsTimeSyncSvcOK(detailed, instanceInfoExtends)
	if err != nil {
		return false, errors.Errorf("validate time sync failed, error:%v", err)
	}

	// check all instance
	for _, instance := range detailed.InstanceInfos {
		os := conf.OS{
			OSName:        instance.OS,
			OSVersion:     instance.OSVersion,
			KernelVersion: instance.Kernel,
			Arch:          instance.Arch,
		}

		if err := c.validateOS(&os); err != nil {
			return false, errors.Errorf("instance: %s validate OS failed, error: %v", instance.PrivateIP, err)
		}

		if err := c.validateInstanceResource(instance.CPU, instance.Memory, instance.SystemDisk, instance.DataDisk); err != nil {
			return false, errors.Errorf("instance: %s validateInstanceResource failed, error: %v", instance.PrivateIP, err)
		}
	}

	return true, nil
}

func (c ClusterValidator) validateHostName(detailed *types.ClusterInfoDetailed) error {
	hostNameMap := make(map[string]string, len(detailed.InstanceInfos))
	for _, instance := range detailed.InstanceInfos {
		if preIP, found := hostNameMap[instance.HostName]; found {
			return errors.Errorf("hostname of %s is duplicate with host %s", instance.PrivateIP, preIP)
		}
		hostNameMap[instance.HostName] = instance.PrivateIP
	}
	return nil
}

func (c ClusterValidator) IsTimeSyncSvcOK(detailed *types.ClusterInfoDetailed, instanceInfoExtends map[string]*types.InstanceInfoExtended) (bool, error) {
	var hasTimeSvcHosts []string
	var notHasTimeSvcHosts []string
	for _, instance := range detailed.InstanceInfos {
		host := instance.PrivateIP

		instanceInfoExtend := instanceInfoExtends[host]
		if instanceInfoExtend.TimeSyncStatus.Ntpd == "active" && instanceInfoExtend.TimeSyncStatus.Chronyd == "active" {
			return false, fmt.Errorf("host %s active ntpd.service and chronyd.service both, please disable one of them")
		}
		timeSvc := ""
		if instanceInfoExtend.TimeSyncStatus.Ntpd == "active" {
			timeSvc = "ntp"
		}
		if instanceInfoExtend.TimeSyncStatus.Chronyd == "active" {
			timeSvc = "chrony"
		}
		if timeSvc != "" {
			hasTimeSvcHosts = append(hasTimeSvcHosts, host)
			//Check time is sync
			output := c.Ssh.Cmd(host, "date +%s")
			if output != nil {
				remoteTime, err := strconv.Atoi(strings.Replace(strings.Replace(string(output), "\r", "", -1), "\n", "", -1))
				if err != nil {
					return false, fmt.Errorf("get remote time of %s failed, error:%s", host, err.Error())
				}

				localTime := time.Now().Unix()
				timediff := int64(remoteTime) - localTime
				if (timediff > 5) || (-5 > timediff) {
					return false, fmt.Errorf("host %s has config %s, but its time diff between master0 greater than 5s", host, timeSvc)
				}
			} else { // command error
				return false, fmt.Errorf("get remote time of %s failed, output is nil", host)
			}
		} else {
			notHasTimeSvcHosts = append(notHasTimeSvcHosts, host)
		}
	}

	if len(hasTimeSvcHosts) == 0 {
		return false, fmt.Errorf("all hosts has no time sync service")
	} else if len(notHasTimeSvcHosts) == 0 {
		return true, nil
	} else {
		return false, fmt.Errorf("some hosts[%s] config time sync service, but some hosts[%s] not, please check",
			strings.Join(hasTimeSvcHosts, ","), strings.Join(notHasTimeSvcHosts, ","))
	}
}

func (c ClusterValidator) validateOS(os *conf.OS) error {
	match := false
	for _, r := range c.SupportedOS {
		if r.OSName == os.OSName &&
			strings.HasPrefix(os.OSVersion, r.OSVersion) &&
			strings.HasPrefix(os.KernelVersion, r.KernelVersion) &&
			r.Arch == os.Arch {
			match = true
			break
		}
	}
	if !match {
		return fmt.Errorf("The current host is: \n%s.\nThe OS only support: \n%s", os.GetStr(), strings.Join(conf.GetSupportedOSStr(), ",\n"))
	}
	return nil
}

func (c ClusterValidator) validateInstanceResource(cpu, memory int32, systemDisk, dataDisks types.DiskSlice) error {

	if cpu < c.HardwareResourceRequired.CpuMinimum {
		return fmt.Errorf("cpu cores should >=%d", c.HardwareResourceRequired.CpuMinimum)
	}

	if memory < c.HardwareResourceRequired.MemMinimum {
		return fmt.Errorf("memory capacity should >=%dGB", c.HardwareResourceRequired.MemMinimum)
	}

	if len(systemDisk) < 1 {
		return fmt.Errorf("systemDisk not found")
	}
	if systemDisk[0].Capacity < c.HardwareResourceRequired.SystemDiskMinimum {
		return fmt.Errorf("systemDisk capacity should >=%dGB", c.HardwareResourceRequired.SystemDiskMinimum)
	}

	return nil
}
