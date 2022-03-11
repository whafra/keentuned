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

type request struct {
	params  []map[string]interface{}
	ip      string
	groupID int
	ipIndex int
	body    interface{}
}

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
		req := group.newRequester(groupID)
		for _, ip := range group.IPs {
			wg.Add(1)
			req.ip = ip
			go func(req request) {
				tuner.apply(&wg, targetFinishStatus, req)
			}(req)
		}
	}

	wg.Wait()

	var errDetail string
	for index, status := range targetFinishStatus {
		applySummary += fmt.Sprintf("\ttarget %v, apply result: %v", index+1, status)
		if strings.Contains(status, "apply failed") {
			errDetail += fmt.Sprintf("\t%v", status)
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

func (tuner *Tuner) apply(wg *sync.WaitGroup, targetFinishStatus []string, req request) {
	start := time.Now()
	var errMsg error
	req.ipIndex = config.KeenTune.IPMap[req.ip]
	config.IsInnerApplyRequests[req.ipIndex] = true
	defer func() {
		wg.Done()
		config.IsInnerApplyRequests[req.ipIndex] = false
		if errMsg != nil {
			targetFinishStatus[req.ipIndex-1] = fmt.Sprintf("target [%v] apply failed, errmsg %v", req.ipIndex, errMsg)
		}
	}()

	var applyResult string
	if tuner.ReadConfigure {
		applyResult, errMsg = tuner.Group[req.groupID].Get(req)
	} else {
		applyResult, errMsg = tuner.Group[req.groupID].Set(req)
	}

	if errMsg != nil {
		return
	}

	targetFinishStatus[req.ipIndex-1] = strings.TrimPrefix(applyResult, "\t")
	tuner.timeSpend.apply += utils.Runtime(start).Count
}

func (gp *Group) Set(req request) (string, error) {
	var setResult string
	for index := range gp.Params {
		if gp.Params[index] == nil {
			continue
		}

		gp.ReadOnly = false
		req.body = gp.applyReq(req.ip, req.params[index])
		result, err := gp.Configure(req)
		if err != nil {
			return result, err
		}

		setResult += result
	}

	return setResult, nil
}

func (gp *Group) Configure(req request) (string, error) {
	uri := fmt.Sprintf("%v:%v/configure", req.ip, gp.Port)
	body, err := http.RemoteCall("POST", uri, req.body)
	if err != nil {
		return "", fmt.Errorf("remote call: %v", err)
	}

	applyResult, paramInfo, err := GetApplyResult(body, req.ipIndex)
	if err != nil {
		return "", fmt.Errorf("apply response: %v", err)
	}

	// pay attention to: the results in the same group are the same and only need to be updated once to prevent map concurrency security problems
	if gp.AllowUpdate[req.ip] {
		err = gp.updateParams(paramInfo)
		if err != nil {
			return applyResult, fmt.Errorf("update apply result: %v", err)
		}

		gp.updateDump(paramInfo)
	}

	return applyResult, nil
}

func (gp *Group) Get(req request) (string, error) {
	gp.ReadOnly = true
	req.body = gp.applyReq(req.ip, gp.MergedParam)
	return gp.Configure(req)
}

func (gp *Group) newRequester(id int) request {
	var data = make([]map[string]interface{}, len(gp.Params))
	for idx, params := range gp.Params {
		data[idx] = deepCopy(params)
	}

	return request{params: data, groupID: id}
}

