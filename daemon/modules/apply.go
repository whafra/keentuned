package modules

import (
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/utils"
	"keentune/daemon/common/utils/http"
	"sync"
	"time"
)

func (tuner *Tuner) Apply() {
	wg := sync.WaitGroup{}
	var errMsg error
	var targetFinishStatus = make(map[int]string, len(tuner.Target.IPs))
	begin := tuner.timeSpend.apply
	for index, ip := range tuner.Target.IPs {
		wg.Add(1)

		go func(index int, ip string) () {
			start := time.Now()
			id := index + 1
			defer func() {
				wg.Done()
				if errMsg != nil {
					targetFinishStatus[id] = fmt.Sprintf("%v", errMsg)
				}
			}()

			host := fmt.Sprintf("%v:%v", ip, config.KeenTune.Ports[index])
			body, err := http.RemoteCall("POST", host+"/configure", tuner.applyReq(index, ip))
			if err != nil {
				errMsg = fmt.Errorf("remote call: %v", err)
				return
			}

			_, _, err = GetApplyResult(body, id)
			if err != nil {
				errMsg = fmt.Errorf("apply response: %v", err)
				return
			}

			// TODO parse dump configure info

			tuner.timeSpend.apply += utils.Runtime(start).Count
			targetFinishStatus[id] = "success"
		}(index, ip)
	}

	wg.Wait()
	fmt.Printf("use time: %.3fs", (tuner.timeSpend.apply - begin).Seconds())
	return
}

func (t *Target) applyResult(status map[int]string, results map[string]Configuration) (string, []Configuration, error) {
	var retConfigs []Configuration
	var retSuccessInfo string
	for index, ip := range t.IPs {
		id := index + 1
		sucInfo, ok := status[id]
		retSuccessInfo += fmt.Sprintf("\n\ttarget id %v, apply result: %v", id, sucInfo)
		if sucInfo != "success" || !ok {
			continue
		}
		retConfigs = append(retConfigs, results[ip])
	}

	if len(retConfigs) == 0 {
		return retSuccessInfo, retConfigs, fmt.Errorf("get target configuration result is null")
	}

	if len(retConfigs) != len(t.IPs) {
		return retSuccessInfo, retConfigs, fmt.Errorf("partial success")
	}

	return retSuccessInfo, retConfigs, nil
}
