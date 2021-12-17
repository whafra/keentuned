package modules

import (
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils/http"
	"fmt"
	"os"
	"strings"
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

func isInterrupted() bool {
	select {
	case <-StopSig:
		rollback()
		return true
	default:
		return false
	}
}

func rollback() {
	url := config.KeenTune.TargetIP + ":" + config.KeenTune.TargetPort + "/rollback"
	if err := http.ResponseSuccess("POST", url, nil); err != nil {
		log.Warnf("", "rollback failed err :%v", err)
	}

	config.IsInnerRequests = false
}

