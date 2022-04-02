package modules

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"keentune/daemon/common/utils/http"
	"reflect"
	"regexp"
	"strings"
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

type DetectResult struct {
	Success bool        `json:"suc"`
	Value   int         `json:"value"`
	Message interface{} `json:"msg"`
}

const defMarcoString = "#!([0-9A-Za-z_]+)#"

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

	if paramName == "Parallel" {
		param["name"] = paramName
		param["domain"] = domainName
		if err := detectParam(param, paramName); err != nil {
			return Parameter{}, fmt.Errorf("detect macro defination param:%v", err)
		}

		delete(param, "name")
		delete(param, "domain")
	}

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

func detectParam(param map[string]interface{}, paramName string) error {
	if paramName != "Parallel" {
		return nil
	}
	ranges, ok := param["range"].([]interface{})
	if !ok {
		return fmt.Errorf("assert range to slice interface failed, real type %v", reflect.TypeOf(param["range"]))
	}
	var range2Int []int
	var detectedMacroValue = make(map[string]int)
	for _, v := range ranges {
		value, ok := v.(float64)
		if ok {
			range2Int = append(range2Int, int(value))
			continue
		}
		macroString, ok := v.(string)
		re, _ := regexp.Compile(defMarcoString)
		macros := utils.RemoveRepeated(re.FindAllString(strings.ReplaceAll(macroString, " ", ""), -1))
		if err := getMacroValue(macros, detectedMacroValue); err != nil {
			return fmt.Errorf("get detect value failed: %v", err)
		}
		calcResult, err := utils.Calculate(convertString(macroString, detectedMacroValue))
		if err != nil {
			return fmt.Errorf("calculate err: %v", err)
		}
		range2Int = append(range2Int, int(calcResult))
	}
	param["range"] = range2Int
	return nil
}
func convertString(macroString string, macroMap map[string]int) string {
	retStr := strings.ReplaceAll(macroString, " ", "")
	for name, value := range macroMap {
		retStr = strings.ReplaceAll(retStr, name, fmt.Sprint(value))
	}
	return retStr
}
func getMacroValue(macros []string, detectedMacroValue map[string]int) error {
	if len(macros) == 0 {
		return nil
	}
	var macroMap = make(map[string]string)
	detectFile := fmt.Sprintf("%v/detect/detect.json", config.KeenTune.Home)
	bytes, err := ioutil.ReadFile(detectFile)
	if err != nil {
		return fmt.Errorf("read detect json file err:%v", err)
	}
	var macroNames []string
	for _, macro := range macros {
		if _, ok := detectedMacroValue[macro]; ok {
			continue
		}
		name := strings.TrimSuffix(strings.TrimPrefix(macro, "#!"), "#")
		lowerName := strings.ToLower(name)
		var tempMap map[string]string
		macroNames = append(macroNames, name)
		err = json.Unmarshal(bytes, &tempMap)
		if err != nil {
			return fmt.Errorf("unmarshal detect json file err:%v", err)
		}
		index := 0
		for key, value := range tempMap {
			index++
			if strings.ToLower(key) == lowerName {
				macroMap[lowerName] = value
				break
			}
			if index == len(tempMap) {
				return fmt.Errorf("detect can't find matched macro :#!%v#", name)
			}
		}
	}
	if len(macroMap) == 0 {
		return nil
	}
	return detect(macroMap, macroNames, detectedMacroValue)
}
func detect(macroMap map[string]string, macroNames []string, detectedMacroValue map[string]int) error {
	requestMap := map[string]interface{}{
		"data": macroMap,
	}
	url := fmt.Sprintf("%v:%v/detect", config.KeenTune.TargetIP[0], config.KeenTune.TargetPort)
	resp, err := http.RemoteCall("POST", url, requestMap)
	var respMap map[string]DetectResult
	err = json.Unmarshal(resp, &respMap)
	if err != nil {
		return fmt.Errorf("unmarshal detect response err:%v", err)
	}
	for _, name := range macroNames {
		result, ok := respMap[strings.ToLower(name)]
		if !ok {
			return fmt.Errorf("get response %v err:%v", name, err)
		}
		if !result.Success {
			return fmt.Errorf("get response %v failed:%v", name, result.Message)
		}
		macro := fmt.Sprintf("#!%v#", name)
		detectedMacroValue[macro] = result.Value
	}
	return nil
}

