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

// StopSig ...
var StopSig chan os.Signal

// Status code
const (
	// SUCCESS status code
	SUCCESS = iota + 1
	WARNING
	FAILED
)

const multiRecordSeparator = "*#++#*"

// backup doesn't exist
const (
	// BackupNotFound error information
	BackupNotFound = "Can not find backup file"
	FileNotExist   = "do not exists"
	NoNeedRollback = "don't need rollback"
	NoBackupFile   = "No backup file was found"
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

// Rollback ...
func Rollback(logName string, callType string) (string, error) {
	tune := new(Tuner)
	tune.logName = logName
	tune.initParams()
	var err error
	if callType == "original" {
		err = tune.original()
	} else {
		err = tune.rollback()
	}

	if err != nil {
		return tune.rollbackFailure, err
	}

	return tune.rollbackDetail, nil
}

func (gp *Group) concurrentSuccess(uri string, request interface{}) (string, bool) {
	wg := sync.WaitGroup{}
	var sucCount = new(int)
	var detailInfo = new(string)
	var failedInfo = new(string)
	unAVLParams := make([]map[string]map[string]string, len(gp.IPs))

	for index, ip := range gp.IPs {
		wg.Add(1)
		id := config.KeenTune.IPMap[ip]
		config.IsInnerApplyRequests[id] = false
		go func(index, groupID int, ip string, wg *sync.WaitGroup) {
			defer wg.Done()
			url := fmt.Sprintf("%v:%v/%v", ip, gp.Port, uri)
			var msg string
			var status int
			if uri != "backup" {
				msg, status = remoteCall("POST", url, request)
			} else {
				unAVLParams[index-1], msg, status = callBackup("POST", url, request)
			}

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

	if uri == "backup" {
		warningInfo, status := gp.deleteUnAVLConf(unAVLParams)
		for _, warn := range strings.Split(warningInfo, ";") {
			parts := strings.Split(warn, "\t")
			if len(parts) != 2 {
				continue
			}
			notMetInfo := fmt.Sprintf(backupENVNotMetFmt, parts[0], parts[1])
			*detailInfo += fmt.Sprintf("%v%v", notMetInfo, multiRecordSeparator)
		}

		if status == FAILED {
			return *detailInfo, false
		}

		if status == WARNING {
			return *detailInfo, true
		}

		return warningInfo, true
	}

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

	return "", parseMsg(response.Msg)
}

func parseMsg(msg interface{}) int {
	switch info := msg.(type) {
	case map[string]interface{}:
		var count int
		for _, value := range info {
			message := fmt.Sprint(value)
			if strings.Contains(message, BackupNotFound) || strings.Contains(message, FileNotExist) ||
				strings.Contains(message, NoNeedRollback) || strings.Contains(message, NoBackupFile) {
				count++
			}
		}

		if count == len(info) && count > 0 {
			return WARNING
		}
		return SUCCESS
	case string:
		if strings.Contains(info, BackupNotFound) || strings.Contains(info, FileNotExist) ||
			strings.Contains(info, NoNeedRollback) || strings.Contains(info, NoBackupFile) {
			return WARNING
		}
		return SUCCESS
	}

	return SUCCESS
}

