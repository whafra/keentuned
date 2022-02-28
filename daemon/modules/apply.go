package modules

import (
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"keentune/daemon/common/utils/http"
	"strings"
	"sync"
	"time"
)

func (tuner *Tuner) getConfigure() error {
	tuner.ReadConfigure = true
	return tuner.configure()
}

func (tuner *Tuner) setConfigure() error {
	tuner.ReadConfigure = false
	return tuner.configure()
}

func (tuner *Tuner) configure() error {
	wg := sync.WaitGroup{}
	var applySummary string
	var targetFinishStatus = make([]string, len(config.KeenTune.IPMap))
	start := time.Now()
	for groupID, group := range tuner.Group {
		for _, ip := range group.IPs {
			wg.Add(1)
			go func(ip string, groupID int) {
				tuner.apply(&wg, targetFinishStatus, ip, groupID)
			}(ip, groupID)
		}
	}

	wg.Wait()

	var errDetail string
	for index, status := range targetFinishStatus {
		applySummary += fmt.Sprintf("\n\ttarget %v, apply result: %v", index+1, status)
		if strings.Contains(status, "apply failed") {
			errDetail += fmt.Sprintf("\n%v", status)
		}
	}

	if errDetail != "" {
		return fmt.Errorf(errDetail)
	}

	tuner.applySummary = applySummary
	timeCost := utils.Runtime(start)
	tuner.timeSpend.apply += timeCost.Count

	if tuner.Verbose && !tuner.ReadConfigure {
		log.Infof(tuner.logName, "[Iteration %v] Apply success, %v", tuner.Iteration, timeCost.Desc)
	}

	return nil
}

func (tuner *Tuner) apply(wg *sync.WaitGroup, targetFinishStatus []string, ip string, groupID int) {
	start := time.Now()
	var errMsg error
	id := config.KeenTune.IPMap[ip]
	defer func() {
		wg.Done()
		if errMsg != nil {
			targetFinishStatus[id-1] = fmt.Sprintf("target [%v] apply failed, errmsg %v", id, errMsg)
		}
	}()

	if tuner.ReadConfigure {
		errMsg = tuner.Group[groupID].Get(ip, id)
	} else {
		errMsg = tuner.Group[groupID].Set(ip, id)
	}

	if errMsg != nil {
		return
	}

	targetFinishStatus[id-1] = "success"
	tuner.timeSpend.apply += utils.Runtime(start).Count
}

func (gp *Group) Set(ip string, id int) error {
	for index := range gp.Params {
		if gp.Params[index] == nil {
			continue
		}

		gp.ReadOnly = false
		err := gp.Configure(ip, id, gp.applyReq(ip, gp.Params[index]))
		if err != nil {
			return err
		}
	}

	return nil
}

func (gp *Group) Configure(ip string, id int, request interface{}) error {
	uri := fmt.Sprintf("%v:%v/configure", ip, gp.Port)
	body, err := http.RemoteCall("POST", uri, request)
	if err != nil {
		return fmt.Errorf("remote call: %v", err)
	}

	_, paramInfo, err := GetApplyResult(body, id)
	if err != nil {
		return fmt.Errorf("apply response: %v", err)
	}

	// pay attention to: the results in the same group are the same and only need to be updated once to prevent map concurrency security problems
	if gp.AllowUpdate[ip] {
		err = gp.updateParams(paramInfo)
		if err != nil {
			return fmt.Errorf("update apply result: %v", err)
		}

		gp.updateDump(paramInfo)
	}

	return nil
}

func (gp *Group) Get(ip string, ipIndex int) error {
	gp.ReadOnly = true
	return gp.Configure(ip, ipIndex, gp.applyReq(ip, gp.MergedParam))
}

