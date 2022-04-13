package modules

import (
	"encoding/json"
	"fmt"
	"keentune/daemon/common/config"
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
	BackupNotFound = "Can not find backup file "
	FileNotExist   = "do not exists"
)

func (tuner *Tuner) isInterrupted() bool {
	select {
	case <-StopSig:
		tuner.rollback()
		return true
	default:
		return false
	}
}

func Rollback(logName string, tune_type string) (string, error) {
	tune := new(Tuner)
	tune.logName = logName
	tune.initParams()
	err := tune.rollback()
	if err != nil {
		return tune.rollbackFailure, err
	}

	return tune.rollbackDetail, nil
}

func Backup(logName string, tune_type string) error {
	tune := new(Tuner)
	tune.logName = logName
	tune.initParams()
	return tune.backup()
}

func (gp *Group) concurrentSuccess(uri string, request interface{}) (string, bool) {
	wg := sync.WaitGroup{}
	var sucCount = new(int)
	var detailInfo = new(string)
	var failedInfo = new(string)

	for index, ip := range gp.IPs {
		wg.Add(1)
		id := config.KeenTune.IPMap[ip]
		config.IsInnerApplyRequests[id] = false
		go func(index, groupID int, ip string, wg *sync.WaitGroup) {
			defer wg.Done()
			url := fmt.Sprintf("%v:%v/%v", ip, gp.Port, uri)
			msg, status := remoteCall("POST", url, request)

			switch status {
			case SUCCESS:
				*sucCount++
			case WARNING:
				*sucCount++
				*detailInfo += fmt.Sprintf("target%v-%v, ", groupID, index)
			case FAILED:
				*failedInfo += fmt.Sprintf("target%v-%v %v; ", groupID, index, msg)
			}

			return
		}(index+1, gp.GroupNo, ip, &wg)
	}

	wg.Wait()
	if *sucCount == len(gp.IPs) {
		return *detailInfo, true
	}

	return *failedInfo, false
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

	msg, ok := response.Msg.(string)
	if ok && (strings.Contains(msg, BackupNotFound) || strings.Contains(msg, FileNotExist)) {
		return "", WARNING
	}

	return "", SUCCESS
}

