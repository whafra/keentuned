package modules

import (
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"keentune/daemon/common/utils/http"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// Configuration define a group of parameter and benchmark score in this configuration.
type Configuration struct {
	Parameters []Parameter           `json:"parameters"`
	Score      map[string]ItemDetail `json:"score"`
	Round      int                   `json:"current_round"`
	budget     float32
	timeSpend  utils.TimeSpend
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
	// aquire API return round is 1 less than the actual round value
	configuration.Round += 1

	err := file.Dump2File(GetTuningWorkPath(fileName), fileName+suffix, configuration)
	if err != nil {
		log.Errorf(log.ParamTune, "dump config info to json file [%v] err: %v", fileName, err)
		return
	}
	return
}

// Load configuration from profile file
func (configuration Configuration) Load() {}

// Apply configuration to Client
func (configuration Configuration) Apply(timeCost *time.Duration) (Configuration, error) {
	start := time.Now()
	host := config.KeenTune.TargetIP + ":" + config.KeenTune.TargetPort
	applyReq, err := configuration.assembleApplyRequestMap()
	if err!=nil {
		return Configuration{}, err
	}

	body, err := http.RemoteCall("POST", host+"/configure", applyReq)
	if err != nil {
		log.Errorf(log.ParamTune, "[Apply]RemoteCall err:[%v]\n", err)
		return Configuration{}, err
	}

	retConfig, err := configuration.parseApplyResponse(body)
	if err != nil {
		return retConfig, err
	}

	retConfig.Round = configuration.Round
	retConfig.timeSpend = utils.Runtime(start)
	*timeCost += retConfig.timeSpend.Count
	return retConfig, nil
}

func (configuration Configuration) assembleApplyRequestMap() (map[string]interface{}, error) {
	domainMap := make(map[string][]map[string]interface{})
	reqApplyMap := make(map[string]interface{})

	//  step 1: assemble domainMap type:map[string][]map[string]interface{}
	for _, param := range configuration.Parameters {
		paramMap, err := utils.Interface2Map(param)
		if err != nil {
			log.Warnf(log.ParamTune, "StructToMap err:[%v]\n", err)
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
			name := param["name"].(string)
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

	retRequst := map[string]interface{}{}
	retRequst["data"] = reqApplyMap
	retRequst["resp_ip"] = respIP
	retRequst["resp_port"] = config.KeenTune.Port	

	return retRequst, nil
}

func (configuration Configuration) parseApplyResponse(body []byte) (Configuration, error) {
	applyResp, err := GetApplyResult(body)
	if err != nil {
		return Configuration{}, err
	}

	var paramCollection = make(map[string]Parameter)
	for domain, param := range applyResp {
		paramMap, ok := param.(map[string]interface{})
		if !ok {
			return Configuration{}, fmt.Errorf("[parseApplyResponse] assert param type[%v] failed", reflect.TypeOf(param))
		}

		for name, orgValue := range paramMap {
			var appliedInfo Parameter
			orgParamMap, ok := orgValue.(map[string]interface{})
			if !ok {
				return Configuration{}, fmt.Errorf("[parseApplyResponse] assert param type[%v] failed", reflect.TypeOf(orgValue))
			}
			err := utils.Map2Struct(orgParamMap, &appliedInfo)
			if err != nil {
				return Configuration{}, fmt.Errorf("parse apply response MapToStruct err:[%v]\n", err)
			}

			appliedInfo.DomainName = domain

			if appliedInfo.Dtype == "string" && strings.Contains(appliedInfo.Value.(string), "\t") {
				appliedInfo.Value = strings.ReplaceAll(appliedInfo.Value.(string), "\t", " ")
			}

			paramCollection[name] = appliedInfo
		}
	}

	for index := range configuration.Parameters {
		paramInfo, ok := paramCollection[configuration.Parameters[index].ParaName]
		if !ok {
			log.Warnf(log.ParamTune, "find [%v] value from apply configure response failed", configuration.Parameters[index].ParaName)
			continue
		}

		configuration.Parameters[index].Value = paramInfo.Value
	}

	return configuration, nil
}

func GetApplyResult(sucBytes []byte) (map[string]interface{}, error) {
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
		Msg     string                 `json:"msg"`
	}
	
	select {
	case body := <-config.ApplyResultChan:
		log.Debugf(log.ParamTune, "apply result :[%v]\n", string(body))
		if err := json.Unmarshal(body, &applyResp); err != nil {
			log.Errorf(log.ParamTune, "parse apply response Unmarshal err:[%v]\n", err)
			return nil, err
		}
	
	}

	if !applyResp.Success {
		return nil, fmt.Errorf("get apply result failed, msg: %v", applyResp.Msg)
	}
	
	return applyResp.Data, nil
}
