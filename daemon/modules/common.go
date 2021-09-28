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

func GetProfileWorkPath(fileName string) string {
	return assembleFilePath(config.KeenTune.DumpConf.DumpHome, "profile", fileName)
}

func GetSensitizePath() string {
	return assembleFilePath(config.KeenTune.DumpConf.DumpHome, "sensitize", "")
}

func GetParamHomePath() string {
	return assembleFilePath(strings.TrimRight(config.KeenTune.Home, "/"), "parameter", "") + "/"
}

func GetProfileHomePath() string {
	return assembleFilePath(strings.TrimRight(config.KeenTune.Home, "/"), "profile", "") + "/"
}


func assembleFilePath(prefix, partition, fileName string) string {
	if fileName == "" {
		return fmt.Sprintf("%s/%s", prefix, partition)
	}

	return fmt.Sprintf("%s/%s/%s", prefix, partition, fileName)
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
		log.Errorf(log.ParamTune, "rollback failed when the system is interrupted, err :%v", err)
	}
}
