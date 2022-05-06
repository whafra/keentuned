package modules

import (
	"encoding/json"
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils/http"
	"os"
	"strings"
	"sync"
)

var StopSig chan os.Signal

const (
	SUCCESS = iota + 1
	WARNING
	FAILED
)

const (
	BackupNotFound = "Can not find backup file"
	FileNotExist   = "do not exists"
)

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

func Rollback(logName string) (string, error) {
	result, allSuc := ConcurrentRequestSuccess(logName, "rollback", nil)
	if !allSuc {
		return result, fmt.Errorf("rollback failed")
	}

	return result, nil
}

func Backup(logName string, request interface{}) (string, bool) {
	return ConcurrentRequestSuccess(logName, "backup", request)
}

func ConcurrentRequestSuccess(logName, uri string, request interface{}) (string, bool) {
	wg := sync.WaitGroup{}
	var sucCount = new(int)
	var warnCount = new(int)
	var detailInfo = new(string)
	var failedInfo = new(string)
	for index, ip := range config.KeenTune.TargetIP {
		wg.Add(1)
		id := index + 1
		config.IsInnerApplyRequests[id] = false
		go func(id int, ip string) () {
			defer wg.Done()
			url := fmt.Sprintf("%v:%v/%v", ip, config.KeenTune.TargetPort, uri)

			msg, status := remoteCall("POST", url, request)
			switch status {
			case SUCCESS:
				*sucCount++
			case WARNING:
				*sucCount++
				*warnCount++
				*detailInfo += fmt.Sprintf("target-%v, ", id)
			case FAILED:
				*failedInfo += fmt.Sprintf("target-%v %v; ", id, msg)
				log.Errorf(logName, "target-%v %v failed: %v", id, uri, msg)
			}
		}(id, ip)
	}

	wg.Wait()

	switch uri {
	case "backup":
		if *sucCount == len(config.KeenTune.TargetIP) {
			return "", true
		}

	case "rollback":
		if *warnCount == len(config.KeenTune.TargetIP) {
			return "No need to Rollback", true
		}

		if *sucCount == len(config.KeenTune.TargetIP) {
			if *detailInfo != "" {
				return fmt.Sprintf("Partial success: %v No Need to Rollback", *detailInfo), true
			}

			return "", true
		}
	}

	return strings.TrimSuffix(*failedInfo, ";\n") + ".", false
}

func remoteCall(method string, url string, request interface{}) (string, int) {
	resp, err := http.RemoteCall(method, url, request)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			return "server is offline", FAILED
		}

		return err.Error(), FAILED
	}

	var response struct {
		Suc bool        `json:"suc"`
		Msg interface{} `json:"msg"`
	}

	if err = json.Unmarshal(resp, &response); err != nil {
		return string(resp), FAILED
	}

	if !response.Suc {
		return fmt.Sprintf("%v", response.Msg), FAILED
	}

	return "", parseMsg(response.Msg)
}

func parseMsg(msg interface{}) int {
	switch info := msg.(type) {
	case map[string]interface{}:
		var count int
		for _, value := range info {
			message := fmt.Sprint(value)
			if strings.Contains(message, BackupNotFound) || strings.Contains(message, FileNotExist) {
				count++
			}
		}

		if count == len(info) && count > 0 {
			return WARNING
		}

		return SUCCESS
	case string:
		if strings.Contains(info, BackupNotFound) || strings.Contains(info, FileNotExist) {
			return WARNING
		}
		return SUCCESS
	}

	return SUCCESS
}

