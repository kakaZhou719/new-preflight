package run

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"preflight/checker/cluster/api/types"
	"preflight/checker/cluster/pkg/conf"
	"preflight/checker/cluster/pkg/parse"
	"preflight/pkg/logger"
	"preflight/pkg/sshcmd/sshutil"
	"preflight/utils"
	"strconv"
	"strings"
)

func ParseClusterInfo(brief *types.ClusterInfoBrief) (clusterInfoDetailed *types.ClusterInfoDetailed, instanceInfoExtends map[string]*types.InstanceInfoExtended, err error) {
	detailed := &types.ClusterInfoDetailed{}
	instanceInfoExtends = make(map[string]*types.InstanceInfoExtended, len(brief.Hosts))

	err = parseInstances(brief.Hosts, conf.SshConfig, detailed, instanceInfoExtends)
	if err != nil {
		return nil, nil, err
	}
	clusterConfigStream, err := yaml.Marshal(detailed)
	if err != nil {

		return nil, nil, err
	}

	err = ioutil.WriteFile("cluster-info-detailed.yaml", clusterConfigStream, 0644)
	if err != nil {

		return nil, nil, err
	}
	return detailed, instanceInfoExtends, err
}

func parseInstances(hosts []string, sshConfig sshutil.SSH, info *types.ClusterInfoDetailed, instanceInfoExtends map[string]*types.InstanceInfoExtended) (err error) {
	if err := parse.DumpScripts(); err != nil {
		return errors.Errorf("save scripts failed: %v", err )
	}

	masterNeed := 3
	identifier := "master"

	checkerPath := parse.GetScriptPath(parse.ParseInstance.Name())
	uuidForDebug := uuid.New()
	localTmpFilePath := filepath.Join(filepath.Dir(checkerPath), fmt.Sprintf("%s_%s.sh", parse.GetScriptNameFromFile(checkerPath), uuidForDebug.String()))
	defer func() {
		if err := os.Remove(localTmpFilePath); err != nil {
			logger.Error(err.Error())
		}
	}()

	if err := util.CopyFile(checkerPath, localTmpFilePath); err != nil {
		return errors.Errorf("copy file %s to %s failed", checkerPath, localTmpFilePath)
	}

	remoteScriptPath := parse.GetTmpScriptPath(parse.ParseInstance.Name())
	if ok := parse.FileDistribution(hosts, localTmpFilePath, remoteScriptPath, sshConfig); ok != true {
		return errors.Errorf("copy file %s to %s:%s failed", localTmpFilePath, hosts, remoteScriptPath)
	}

	for i, host := range hosts {
		if sshConfig.IsFileExist(host, remoteScriptPath) {
			output := sshConfig.Cmd(host, fmt.Sprintf("bash %s %s", remoteScriptPath, parse.ParseInstance.Params()))
			if output != nil {
				str := string(output)
				instanceInfo, instanceInfoExtend, err := parseInstanceInfo(str)
				if err != nil {
					return errors.Errorf("failed to execut command %s:%s failed", hosts, remoteScriptPath)

				}
				if i >= masterNeed {
					identifier = "worker"
				}
				instanceInfo.PrivateIP = IPFormat(host)
				instanceInfo.Identifier = identifier

				info.InstanceInfos = append(info.InstanceInfos, *instanceInfo)
				instanceInfoExtends[IPFormat(host)] = instanceInfoExtend
			} else { // command error

			}
		} else {
			return errors.Errorf("[%s]script %s is not found", host, remoteScriptPath)
		}
	}
	return nil
}

func parseInstanceInfo(str string) (*types.InstanceInfo, *types.InstanceInfoExtended, error) {
	var instanceInfoExtended types.InstanceInfoExtended

	str, err := getBetweenStr(str, "##INSTANCE_INFO_BEGIN##", "##INSTANCE_INFO_END##")
	if err != nil {
		return nil, nil, err
	}

	if err := json.Unmarshal([]byte(str), &instanceInfoExtended); err != nil {
		return nil, nil, fmt.Errorf("Unmarshal err, %v\n", err)
	}
	instanceInfo := instanceInfoExtended.InstanceInfo
	instanceInfo.Memory = (instanceInfo.Memory + 1024*1024 - 1) / (1024 * 1024)

	networkDevicesStr, err := base64.StdEncoding.DecodeString(instanceInfoExtended.NetworkDevicesStr)
	if err != nil {
		return nil, nil, err
	}
	for _, netDevStr := range strings.Split(string(networkDevicesStr), "##SPLITER##") {
		netCard := parseNetCard(netDevStr)
		//skip lo,gw
		if netCard.Name == "lo" || netCard.Name == "gw" {
			continue
		}
		instanceInfo.NetworkCards = append(instanceInfo.NetworkCards, &netCard)
		if netCard.IP == instanceInfo.PrivateIP {
			instanceInfo.MACAddress = netCard.MAC
		}
	}

	blockDevicesStr, err := base64.StdEncoding.DecodeString(instanceInfoExtended.BlockDevicesStr)
	if err != nil {
		return nil, nil, err
	}
	systemDiskName := ""
	disks := make(map[string]*types.Disk)
	for _, blkDevStr := range strings.Split(string(blockDevicesStr), "##SPLITER##") {
		blkDev, parentName, err := parseBlkDev(blkDevStr)
		if err != nil {
			return nil, nil, err
		}

		if isSystemDisk(blkDev) {
			if parentName != "" {
				systemDiskName = parentName
			} else {
				systemDiskName = blkDev.Name
			}
		}
		disks[blkDev.Name] = blkDev
	}
	instanceInfo.SystemDisk = types.DiskSlice{disks[systemDiskName]}
	delete(disks, systemDiskName)
	for _, value := range disks {
		if value.Type == "disk" {
			instanceInfo.DataDisk = append(instanceInfo.DataDisk, value)
		}
	}

	return &instanceInfo, &instanceInfoExtended, nil
}

func parseNetCard(netDevStr string) types.NetWorkCard {
	netInfoMap := make(map[string]string)
	for _, pairStr := range strings.Split(netDevStr, " ") {
		pair := strings.Split(pairStr, "=")
		if len(pair) != 2 {
			logger.Error("parse net device pair: %s failed", pair)
			continue
		}
		value := pair[1]
		if value != "" {
			if value[0] == '"' {
				value = value[1:]
			}
			if value[len(value)-1] == '"' {
				value = value[:len(value)-1]
			}
		}
		netInfoMap[pair[0]] = value
	}
	return types.NetWorkCard{
		Name: netInfoMap["NAME"],
		IP:   netInfoMap["IP"],
		MAC:  netInfoMap["MAC"],
	}
}

func parseBlkDev(blkDevStr string) (*types.Disk, string, error) {
	blkInfoMap := make(map[string]string)
	for _, pairStr := range strings.Split(blkDevStr, " ") {
		pair := strings.Split(pairStr, "=")
		if len(pair) != 2 {
			return &types.Disk{}, "", fmt.Errorf("parse block device pair: %s failed", pair)
		}
		value := pair[1]
		if value != "" {
			if value[0] == '"' {
				value = value[1:]
			}
			if value[len(value)-1] == '"' {
				value = value[:len(value)-1]
			}
		}
		blkInfoMap[pair[0]] = value
	}

	capacityBit, err := strconv.Atoi(blkInfoMap["SIZE"])
	if err != nil {
		return &types.Disk{}, "", fmt.Errorf("block size: %s can't be convert to int", blkInfoMap["SIZE"])
	}

	return &types.Disk{
		Name:       blkInfoMap["NAME"],
		MountPoint: blkInfoMap["MOUNTPOINT"],
		FSType:     blkInfoMap["FSTYPE"],
		Capacity:   int32(capacityBit / (1024 * 1024 * 1024)),
		Type:       blkInfoMap["TYPE"],
	}, blkInfoMap["PKNAME"], nil
}

func getBetweenStr(str, begin, end string) (string, error) {
	n := strings.Index(str, begin)
	if n == -1 {
		return "", fmt.Errorf("can't find begin str")
	}

	m := strings.Index(str, end)
	if m == -1 {
		return "", fmt.Errorf("can't find end str")
	}

	return str[n+len(begin) : m], nil
}

func isSystemDisk(blkdev *types.Disk) bool {
	return blkdev.MountPoint == "/"
}

func IPFormat(host string) string {
	if strings.IndexRune(host, ':') < 0 {
		return host
	}
	return strings.Split(host, ":")[0]
}
