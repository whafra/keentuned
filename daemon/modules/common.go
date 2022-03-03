package modules

import (
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/utils/http"
	"os"
	"strings"
	"sync"
)

var StopSig chan os.Signal

func (tuner *Tuner) isInterrupted() bool {
	select {
	case <-StopSig:
		tuner.rollback()
		return true
	default:
		return false
	}
}

func Rollback(logName string, tune_type string) error {
	tune := new(Tuner)
	tune.logName = logName
	tune.initParams()
	return tune.rollback()
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

	for index, ip := range gp.IPs {
		wg.Add(1)
		id := config.KeenTune.IPMap[ip]
		config.IsInnerApplyRequests[id] = false
		go func(index, id int, ip string, wg *sync.WaitGroup) {
			defer wg.Done()
			url := fmt.Sprintf("%v:%v/%v", ip, gp.Port, uri)
			if err := http.ResponseSuccess("POST", url, request); err != nil {
				*detailInfo += fmt.Sprintf("target [%v] %v;\n", id, err)
				return
			}

			*sucCount++
			*detailInfo += fmt.Sprintf("target [%v] success;\n", id)
		}(index, id, ip, &wg)
	}

	wg.Wait()
	if *sucCount == len(gp.IPs) {
		return *detailInfo, true
	}

	return strings.TrimSuffix(*detailInfo, ";\n") + ".", false
}
