package parse

import (
	"encoding/base64"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"preflight/checker/cluster/pkg/conf"
	"preflight/pkg/logger"
	"preflight/pkg/sshcmd/md5sum"
	"preflight/pkg/sshcmd/sshutil"
	"preflight/utils"
)

var DefaultCheckersMap = make(map[string]Checker)

func init() {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	conf.ScriptDir = filepath.Join(home, ".cluster-checker/scripts/")
	for _, checker := range DefaultCheckers {
		DefaultCheckersMap[checker.Name()] = checker
	}
}

func FileDistribution(hosts []string, srcFilePath, dstFilePath string, sshConfig sshutil.SSH) (succeed bool) {
	var wm sync.WaitGroup
	succeed = true
	for k, v := range sshConfig.Env {
		if err := util.InsertStringToFile(srcFilePath, fmt.Sprintf("%s=%s\n", k, v), 0); err != nil {
			logger.Error(err.Error())
		}
	}
	for _, host := range hosts {
		wm.Add(1)
		go func(host string) {
			defer wm.Done()
			md5 := md5sum.FromLocal(srcFilePath)
			if ok := sshConfig.CopyForMD5(host, srcFilePath, dstFilePath, md5); ok {
				logger.Debug("[%s]copy file md5 validate success", host)
			} else {
				logger.Error("[%s]copy file md5 validate failed", host)
				succeed = false
			}
		}(host)
	}
	wm.Wait()
	return succeed
}

func GetScriptNameFromFile(scriptFilePath string) string {
	checkerFileName := strings.ToLower(filepath.Base(scriptFilePath))
	checkerName := strings.TrimSuffix(checkerFileName, filepath.Ext(checkerFileName))
	return checkerName
}

func DumpScripts() error {
	// 存在目录则不做任何操作
	_, err := os.Stat(conf.ScriptDir)
	if err != nil && !os.IsNotExist(err) {
		return err
	} else if os.IsNotExist(err) {
		err = os.MkdirAll(conf.ScriptDir, os.ModePerm)
		if err != nil {
			logger.Error("create default script dir failed, please create it by your self mkdir -p %s\n", conf.ScriptDir)
			return err
		}
	}

	for _, checker := range DefaultCheckers {
		scriptData, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(checker.Script(), " ", "\n"))
		if err != nil {
			logger.Error("decode script error: %s", err.Error())
			return err
		}
		script := string(scriptData)
		fileName := GetScriptPath(checker.Name())
		if err := ioutil.WriteFile(fileName, []byte(script), 0755); err != nil {
			logger.Error("write to file %s error: %s", fileName, err)
			return err
		}
	}

	return nil
}

func GetScriptPath(checkerName string) string {
	return filepath.Join(conf.ScriptDir, fmt.Sprintf("%s.sh", strings.ToLower(checkerName)))
}

func GetTmpScriptPath(checkerName string) string {
	return fmt.Sprintf("/tmp/%s.sh", strings.ToLower(checkerName))
}
