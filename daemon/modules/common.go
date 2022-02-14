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
	return ConcurrentRequestSuccess(logName, "Backup", request)
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
func (t *Target) Backup() (string, bool) {
	return t.concurrentRequestSuccess("Backup", true)
}

func (t *Target) Rollback() (string, bool) {
	return t.concurrentRequestSuccess("Backup", false)
}

func (t *Target) concurrentRequestSuccess(uri string, needParam bool) (string, bool) {
	wg := sync.WaitGroup{}
	var sucCount int
	var detailInfo string
	var request interface{}
	if (len(t.IPs) != len(t.Params)) || (len(t.IPs) != len(config.KeenTune.Ports)) {
		return "ip length is not matched to param length", false
	}

	for index, ip := range t.IPs {
		wg.Add(1)
		id := index + 1
		config.IsInnerApplyRequests[id] = false
		go func(index, id int, ip string) () {
			defer wg.Done()
			if needParam {
				request = t.Params[index]
			}

			url := fmt.Sprintf("%v:%v/%v", ip, config.KeenTune.Ports[index], uri)
			if err := http.ResponseSuccess("POST", url, request); err != nil {
				detailInfo += fmt.Sprintf("target [%v] %v;\n", id, err)
				return
			}

			sucCount++
		}(index, id, ip)
	}

	wg.Wait()
	if sucCount == len(t.IPs) {
		return "", true
	}

	return strings.TrimSuffix(detailInfo, ";\n") + ".", false
}

