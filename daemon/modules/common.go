package modules

import (
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils/http"
	"os"
	"strings"
	"sync"
)

var StopSig chan os.Signal

func GetTuningWorkPath(fileName string) string {
	return assembleFilePath(config.KeenTune.DumpConf.DumpHome, "parameter", fileName)
}

func GetGenerateWorkPath(fileName string) string {
	return assembleFilePath(config.KeenTune.DumpConf.DumpHome, "parameter/generate", fileName)
}

func GetBenchHomePath() string {
	return assembleFilePath(config.KeenTune.Home, "benchmark", "")
}

func GetProfileWorkPath(fileName string) string {
	return assembleFilePath(config.KeenTune.DumpConf.DumpHome, "profile", fileName)
}

func GetSensitizePath() string {
	return assembleFilePath(config.KeenTune.DumpConf.DumpHome, "sensitize", "")
}

func GetParamHomePath() string {
	return assembleFilePath(config.KeenTune.Home, "parameter", "") + "/"
}

func GetProfileHomePath(fileName string) string {
	if fileName == "" {
		return fmt.Sprintf("%s/%s", config.KeenTune.Home, "profile") + "/"
	}

	return assembleFilePath(config.KeenTune.Home, "profile", fileName)
}

func GetDumpCSVPath() string {
	return assembleFilePath(config.KeenTune.DumpConf.DumpHome, "csv", "")
}

func assembleFilePath(prefix, partition, fileName string) string {
	if fileName == "" {
		return fmt.Sprintf("%s/%s", prefix, partition)
	}
	
	// absolute path
	if strings.HasPrefix(fileName, "/") && strings.Count(fileName, "/") > 1 {
		return fileName
	}

	// relative path
	if strings.Contains(fileName, fmt.Sprintf("%v/", partition)) {
		parts := strings.Split(fileName, fmt.Sprintf("%v/", partition))
		return fmt.Sprintf("%s/%s/%s", prefix, partition, parts[len(parts)-1])
	}

	// file
	return fmt.Sprintf("%s/%s/%s", prefix, partition, strings.TrimPrefix(fileName, "/"))
}

func isInterrupted(logName string) bool {
	select {
	case <-StopSig:
		Rollback(logName)
		return true
	default:
		return false
	}
}

func Rollback(logName string) (string, bool) {
	return ConcurrentRequestSuccess(logName, "rollback", nil)
}

func Backup(logName string, request interface{}) (string, bool) {
	return ConcurrentRequestSuccess(logName, "backup", request)
}

func ConcurrentRequestSuccess(logName, uri string, request interface{}) (string, bool) {
	wg := sync.WaitGroup{}
	var sucCount int
	var detailInfo string
	for index, ip := range config.KeenTune.TargetIP {
		wg.Add(1)
		id := index + 1
		config.IsInnerApplyRequests[id] = false
		go func(id int, ip string) () {
			defer wg.Done()
			url := fmt.Sprintf("%v:%v/%v", ip, config.KeenTune.TargetPort, uri)
			if err := http.ResponseSuccess("POST", url, request); err != nil {
				detailInfo += fmt.Sprintf("target [%v] %v;\n", id, err)
				log.Errorf(logName, "target [%v] %v failed: %v", id, uri, err)
				return
			}

			sucCount++
		}(id, ip)
	}

	wg.Wait()
	if sucCount == len(config.KeenTune.TargetIP) {
		return "", true
	}

	return strings.TrimSuffix(detailInfo, ";\n") + ".", false
}

