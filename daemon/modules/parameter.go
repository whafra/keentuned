package modules

import (
	"fmt"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
)

// Parameter define a os parameter value scope, operating command and value
type Parameter struct {
	DomainName string        `json:"domain"`
	ParaName   string        `json:"name"`
	SetCMD     string        `json:"cmd_set,omitempty"`
	GetCMD     string        `json:"cmd_get,omitempty"`
	Continuity bool          `json:"-"`
	Scope      []interface{} `json:"range,omitempty"`
	Options    []string      `json:"options,omitempty"`
	Sequence   []interface{} `json:"sequence,omitempty"`
	Dtype      string        `json:"dtype"`
	Value      interface{}   `json:"value,omitempty"`
	Base       interface{}   `json:"base,omitempty"`
	Msg        string        `json:"msg,omitempty"`
	Step       int           `json:"step,omitempty"`
	Weight     float32       `json:"weight,omitempty"`
	Success    bool          `json:"suc,omitempty"`
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
func AssembleParams(userParam map[string]map[string]interface{}, totalParamMap ...map[string]map[string]interface{}) (*Configuration, map[string]interface{}) {
	var readParams []Parameter
	var applyInitConfig = new(Configuration)
	var initRequestParam = make(map[string]interface{})
	var initParamSlice = make([]map[string]interface{}, 0)

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

			param, initParam, readParam, err := doAssembleParam(param, compareMap, domainName, name)
			if err != nil {
				log.Warnf("", "do assemble parameters err:%v", err)
				continue
			}

			initParamSlice = append(initParamSlice, initParam)
			readParams = append(readParams, readParam)
			domainMap[name] = param
		}

		userParam[domainName] = domainMap
	}

	applyInitConfig.Parameters = readParams
	initRequestParam["parameters"] = initParamSlice

	return applyInitConfig, initRequestParam
}

func doAssembleParam(originMap, compareMap map[string]interface{}, domainName, paramName string) (map[string]interface{}, map[string]interface{}, Parameter, error) {
	if compareMap == nil {
		return assembleParam(originMap, domainName, paramName)
	}

	var compareParamResult map[string]interface{}
	compareParamResult, err := modifyParam(originMap, compareMap, domainName, paramName)
	if err != nil {
		return compareParamResult, nil, Parameter{}, err
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

//  assembleParam assemble return[0] for init API request paramMap ; return[1] for apply API baseline config params which value is empty only for read
func assembleParam(param map[string]interface{}, domainName, paramName string) (map[string]interface{}, map[string]interface{}, Parameter, error) {
	//  delete `desc` field, not used in any request body
	delete(param, "desc")
	param["name"] = paramName
	param["domain"] = domainName

	var readParam Parameter
	readParamMap := make(map[string]interface{})
	initParamMap := make(map[string]interface{})
	for name, value := range param {
		readParamMap[name] = value
		initParamMap[name] = value
	}

	/* delete `options` 、`range` field, not used in apply configuration api request body */
	delete(readParamMap, "options")
	delete(readParamMap, "range")
	delete(readParamMap, "step")
	err := utils.Map2Struct(readParamMap, &readParam)
	if err != nil {
		return param, initParamMap, Parameter{}, fmt.Errorf("assembleReadParam map to struct err:%v", err)
	}

	readParam.DomainName = domainName
	readParam.ParaName = paramName
	readParam.Value = ""
	readParam.Dtype = "read"
	delete(initParamMap, "value")
	delete(param, "name")
	delete(param, "domain")

	return param, initParamMap, readParam, nil
}
