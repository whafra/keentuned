package modules

import (
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
)

// Parameter define a os parameter value scope, operating command and value
type Parameter struct {
	DomainName string        `json:"domain"`
	ParaName   string        `json:"name"`
	SetCMD     string        `json:"cmd_set,omitempty"`
	GetCMD     string        `json:"cmd_get,omitempty"`
	Scope      []interface{} `json:"range,omitempty"`
	Options    []string      `json:"options,omitempty"`
	Sequence   []interface{} `json:"sequence,omitempty"`
	Dtype      string        `json:"dtype"`
	Value      interface{}   `json:"value,omitempty"`
	Msg        string        `json:"msg,omitempty"`
	Step       int           `json:"step,omitempty"`
	Weight     float32       `json:"weight,omitempty"`
	Success    bool          `json:"suc,omitempty"`
	Base       interface{}   `json:"base,omitempty"`
}

// updateParameter update the partial param by the total param
func updateParameter(partial, total *Parameter) {
	if partial.Dtype == "" {
		partial.Dtype = total.Dtype
	}

	if partial.GetCMD == "" {
		partial.GetCMD = total.GetCMD
	}

	if partial.SetCMD == "" {
		partial.SetCMD = total.SetCMD
	}

	if len(partial.Options) == 0 {
		partial.Options = total.Options
	}

	if partial.Value == nil {
		partial.Value = total.Value
	}

	if len(partial.Scope) == 0 {
		partial.Scope = total.Scope
	}

	if partial.Step == 0 {
		partial.Step = total.Step
	}
}

// AssembleParams assemble params for tune init, include init  and apply  API request params
func AssembleParams(userParam config.DBLMap, totalParamMap ...config.DBLMap) *Configuration {
	var initParams []Parameter
	var initConfig = new(Configuration)

	for domainName, domainMap := range userParam {
		var compareMap map[string]interface{}
		if len(totalParamMap) > 0 {
			compareMap = utils.Parse2Map(domainName, totalParamMap[0])
			if compareMap == nil {
				log.Warnf("", "find param domain [%v] does not exist when modify user params by parsing local total param file", domainName)
			}
		}

		for name, paramValue := range domainMap {
			param, ok := paramValue.(map[string]interface{})
			if !ok {
				log.Warnf("", "parse [%+v] type to map failed", paramValue)
				continue
			}

			initParam, err := doAssembleParam(param, compareMap, domainName, name)
			if err != nil {
				log.Warnf("", "do assemble parameters err:%v", err)
				continue
			}

			initParams = append(initParams, initParam)
			domainMap[name] = param
		}

		userParam[domainName] = domainMap
	}

	initConfig.Parameters = initParams

	return initConfig
}

func doAssembleParam(originMap, compareMap map[string]interface{}, domainName, paramName string) (Parameter, error) {
	if compareMap == nil {
		return assembleParam(originMap, domainName, paramName)
	}

	var compareParamResult map[string]interface{}
	compareParamResult, err := modifyParam(originMap, compareMap, domainName, paramName)
	if err != nil {
		return Parameter{}, err
	}

	return assembleParam(compareParamResult, domainName, paramName)
}

func modifyParam(originMap, compareMap map[string]interface{}, domainName, paramName string) (map[string]interface{}, error) {
	var userParam, totalParam Parameter
	err := utils.Map2Struct(originMap, &userParam)
	if err != nil {
		return nil, fmt.Errorf("modifyParam map to struct err:%v", err)
	}

	if err = utils.Parse2Struct(paramName, compareMap, &totalParam); err != nil {
		return nil, fmt.Errorf("modifyParam parse compareMap %+v to struct err:%v", compareMap, err)
	}

	updateParameter(&userParam, &totalParam)
	resultParamMap, err := utils.Struct2Map(userParam)
	if err != nil {
		return nil, fmt.Errorf("modifyParam struct %+v To map  err:%v", userParam, err)

	}

	return resultParamMap, nil
}

//  assembleParam assemble return[0] for apply API baseline config
func assembleParam(param map[string]interface{}, domainName, paramName string) (Parameter, error) {
	//  delete `desc` field, not used in any request body
	delete(param, "desc")

	var initParam Parameter
	initParamMap := make(map[string]interface{})
	for name, value := range param {
		initParamMap[name] = value
	}

	err := utils.Map2Struct(initParamMap, &initParam)
	if err != nil {
		return Parameter{}, fmt.Errorf("assembleReadParam map to struct err:%v", err)
	}

	initParam.DomainName = domainName
	initParam.ParaName = paramName
	return initParam, nil
}
