package modules

import (
	"encoding/json"
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"keentune/daemon/common/utils/http"
	"reflect"
	"strings"
	"sync"
	"time"
)

// Configuration define a group of parameter and benchmark score in this configuration.
type Configuration struct {
	Parameters []Parameter           `json:"parameters"`
	Score      map[string]ItemDetail `json:"score"`
	Round      int                   `json:"current_round"`
	budget     float32
	timeSpend  utils.TimeSpend
	targetIP   []string
}

type ReceivedConfigure struct {
	Candidate []Parameter           `json:"candidate"`
	Score     map[string]ItemDetail `json:"score,omitempty"`
	Iteration int                   `json:"iteration"`
	Budget    float32               `json:"budget"`
}

type ItemDetail struct {
	Value    float32 `json:"value"`
	Negative bool    `json:"negative"`
	Weight   float32 `json:"weight"`
	Strict   bool    `json:"strict"`
	Baseline float32 `json:"baseline"`
}

// Dump configuration to profile file
func (configuration Configuration) Dump(fileName, suffix string) {
	// acquire API return round is 1 less than the actual round value
	configuration.Round += 1

	err := file.Dump2File(GetTuningWorkPath(fileName), fileName+suffix, configuration)
	if err != nil {
		log.Warnf("", "dump config info to json file [%v] err: %v", fileName, err)
		return
	}
	return
}

// Apply configuration to Client
func (configuration Configuration) Apply(timeCost *time.Duration) (string, []Configuration, error) {
	start := time.Now()
	configuration.targetIP = config.KeenTune.TargetIP
	applyReq, err := configuration.assembleApplyRequestMap()
	if err != nil {
		return "", []Configuration{}, err
	}
	wg := sync.WaitGroup{}
	var errMsg error
	var targetFinishStatus = make(map[int]string, len(configuration.targetIP))
	var applyResults = make(map[string]Configuration, len(configuration.targetIP))
	for index, ip := range configuration.targetIP {
		wg.Add(1)

		go func(id int, ip string) () {
			defer func() {
				wg.Done()
				if errMsg != nil {
					targetFinishStatus[id] = fmt.Sprintf("apply failed, details：%v", errMsg)
				}
			}()

			host := ip + ":" + config.KeenTune.TargetPort
			body, err := http.RemoteCall("POST", host+"/configure", utils.ConcurrentSecurityMap(applyReq, []string{"target_id"}, []interface{}{id}))
			if err != nil {
				errMsg = fmt.Errorf("target [%v] remote call configure: %v", id, err)
				return
			}

			tempResult, err := configuration.parseApplyResponse(body, id)
			if err != nil {
				errMsg = fmt.Errorf("target [%v] parse response err: %v", id, err)
				return
			}

			tempResult.Round = configuration.Round
			tempResult.timeSpend = utils.Runtime(start)
			*timeCost += tempResult.timeSpend.Count
			targetFinishStatus[id] = "success"
			applyResults[ip] = tempResult
		}(index+1, ip)
	}

	wg.Wait()

	return configuration.applyResult(targetFinishStatus, applyResults)
}

func (configuration Configuration) applyResult(status map[int]string, results map[string]Configuration) (string, []Configuration, error) {
	var retConfigs []Configuration
	var retSuccessInfo string
	for index, ip := range configuration.targetIP {
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

	if len(retConfigs) != len(configuration.targetIP) {
		return retSuccessInfo, retConfigs, fmt.Errorf("partial success")
	}

	return retSuccessInfo, retConfigs, nil
}

func (configuration Configuration) assembleApplyRequestMap() (map[string]interface{}, error) {
	domainMap := make(map[string][]map[string]interface{})
	reqApplyMap := make(map[string]interface{})

	//  step 1: assemble domainMap type:map[string][]map[string]interface{}
	for _, param := range configuration.Parameters {
		paramMap, err := utils.Interface2Map(param)
		if err != nil {
			log.Warnf("", "StructToMap err:[%v]\n", err)
			continue
		}
		/* delete `domain` field, not used in apply api request body */
		delete(paramMap, "domain")
		delete(paramMap, "step")
		domainMap[param.DomainName] = append(domainMap[param.DomainName], paramMap)
	}

	// step 2: range the domainMap and change the []map[string]interface{} to map[string]interface{} by key
	for domain, params := range domainMap {
		var tempDomainMap = make(map[string]map[string]interface{})
		for _, param := range params {
			name, ok := param["name"].(string)
			if !ok {
				return nil, fmt.Errorf("%+v get name failed", param)
			}

			/* delete `name` field, not used in apply api request body */
			delete(param, "name")
			tempDomainMap[name] = param
		}

		reqApplyMap[domain] = tempDomainMap
	}

	respIP, err := utils.GetExternalIP()
	if err != nil {
		return nil, fmt.Errorf("run benchmark get real keentuned ip err: %v", err)
	}

	retRequest := map[string]interface{}{}
	retRequest["data"] = reqApplyMap
	retRequest["resp_ip"] = respIP
	retRequest["resp_port"] = config.KeenTune.Port

	return retRequest, nil
}

func (configuration Configuration) parseApplyResponse(body []byte, id int) (Configuration, error) {
	_, paramCollection, err := GetApplyResult(body, id)
	if err != nil {
		return Configuration{}, err
	}

	for index := range configuration.Parameters {
		paramInfo, ok := paramCollection[configuration.Parameters[index].ParaName]
		if !ok {
			log.Warnf("", "find [%v] value from apply configure response failed", configuration.Parameters[index].ParaName)
			continue
		}

		configuration.Parameters[index].Value = paramInfo.Value
	}

	return configuration, nil
}

// collectParam collect param change map to struct map and state param success information
func collectParam(applyResp map[string]interface{}) (string, map[string]Parameter, error) {
	var paramCollection = make(map[string]Parameter)
	var sucCount, failedCount int
	var failedInfoSlice [][]string

	if len(applyResp) > 0 {
		failedInfoSlice = append(failedInfoSlice, []string{"param name", "failed reason"})
	}

	for domain, param := range applyResp {
		paramMap, ok := param.(map[string]interface{})
		if !ok {
			return "", paramCollection, fmt.Errorf("collect Param assert type [%v] to map failed", reflect.TypeOf(param))
		}

		for name, orgValue := range paramMap {
			var appliedInfo Parameter
			err := utils.Map2Struct(orgValue, &appliedInfo)
			if err != nil {
				return "", paramCollection, fmt.Errorf("collect Param:[%v]\n", err)
			}

			appliedInfo.DomainName = domain

			if appliedInfo.Dtype == "string" && strings.Contains(appliedInfo.Value.(string), "\t") {
				appliedInfo.Value = strings.ReplaceAll(appliedInfo.Value.(string), "\t", " ")
			}

			paramCollection[name] = appliedInfo

			if appliedInfo.Success {
				sucCount++
				continue
			}

			failedCount++
			failedInfoSlice = append(failedInfoSlice, []string{name, appliedInfo.Msg})
		}
	}

	var setResult string

	if failedCount == 0 {
		setResult = fmt.Sprintf("total param %v, successed %v, failed %v.", sucCount, sucCount, failedCount)
		return setResult, paramCollection, nil
	}

	setResult = fmt.Sprintf("total param %v, successed %v, failed %v; the failed details:%s", sucCount+failedCount, sucCount, failedCount, utils.FormatInTable(failedInfoSlice))
	return setResult, paramCollection, nil
}

func getApplyResult(sucBytes []byte, id int) (map[string]interface{}, error) {
	var applyShortRet struct {
		Success bool `json:"suc"`
	}

	err := json.Unmarshal(sucBytes, &applyShortRet)
	if err != nil {
		return nil, err
	}

	if !applyShortRet.Success {
		return nil, fmt.Errorf("apply short return failed")
	}

	var applyResp struct {
		Success bool                   `json:"suc"`
		Data    map[string]interface{} `json:"data"`
		Msg     interface{}            `json:"msg"`
	}

	config.IsInnerApplyRequests[id] = true
	select {
	case body := <-config.ApplyResultChan[id]:
		log.Debugf(log.ParamTune, "target id: %v receive apply result :[%v]\n", id, string(body))
		if err := json.Unmarshal(body, &applyResp); err != nil {
			return nil, fmt.Errorf("Parse apply response Unmarshal err: %v", err)
		}

	}

	if !applyResp.Success {
		return nil, fmt.Errorf("get apply result failed, msg: %v", applyResp.Msg)
	}

	return applyResp.Data, nil
}

func GetApplyResult(body []byte, id int) (string, map[string]Parameter, error) {
	applyResp, err := getApplyResult(body, id)
	if err != nil {
		return "", nil, err
	}

	return collectParam(applyResp)
}

