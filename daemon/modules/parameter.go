package modules

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"keentune/daemon/common/utils/http"
	"regexp"
	"strings"
)

// Parameter define a os parameter value scope, operating command and value
type Parameter struct {
	DomainName string        `json:"domain,omitempty"`
	ParaName   string        `json:"name,omitempty"`
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

// UpdateParams update params by total param
func UpdateParams(userParam config.DBLMap) {
	for domainName, domainMap := range userParam {
		fileName := fmt.Sprintf("%s/parameter/%s.json", config.KeenTune.Home, domainName)

		totalParamMap, err := file.ReadFile2Map(fileName)
		if err != nil {
			log.Warnf("", "Read file: '%v' , err: %v\n", fileName, err)
			continue
		}

		compareMap := utils.Parse2Map(domainName, totalParamMap)
		if len(totalParamMap) == 0 || len(compareMap) == 0 {
			log.Warnf("", "domain [%v] does not exist when update params by parsing '%v'", domainName, fileName)
			continue
		}

		for name, paramValue := range domainMap {
			param, ok := paramValue.(map[string]interface{})
			if !ok {
				log.Warnf("", "parse [%+v] type to map failed", paramValue)
				continue
			}

			err := modifyParam(&param, compareMap, name)
			if err != nil {
				log.Warnf("", "modify parameters err:%v", err)
				continue
			}

			domainMap[name] = param
		}

		userParam[domainName] = domainMap
	}

	return
}

func modifyParam(originMap *map[string]interface{}, compareMap map[string]interface{}, paramName string) error {
	var userParam, totalParam Parameter
	err := utils.Map2Struct(originMap, &userParam)
	if err != nil {
		return fmt.Errorf("modifyParam map to struct err:%v", err)
	}

	if err = utils.Parse2Struct(paramName, compareMap, &totalParam); err != nil {
		return fmt.Errorf("modifyParam parse compareMap %+v to struct err:%v", compareMap, err)
	}

	updateParameter(&userParam, &totalParam)
	*originMap, err = utils.Struct2Map(userParam)
	if err != nil {
		return fmt.Errorf("modifyParam struct %+v To map  err:%v", userParam, err)

	}

	return nil
}

func detectParam(param *Parameter) error {
	if len(param.Scope) > 0 {
		var range2Int []interface{}
		var detectedMacroValue = make(map[string]int)
		for _, v := range param.Scope {
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
				return fmt.Errorf("'%v' calculate range err: %v", param.ParaName, err)
			}
			range2Int = append(range2Int, int(calcResult))
		}
		param.Scope = range2Int
	}

	if len(param.Options) > 0 {
		var newOptions []string
		var detectedMacroValue = make(map[string]int)
		for _, v := range param.Options {
			re, _ := regexp.Compile(defMarcoString)
			if !re.MatchString(v) {
				newOptions = append(newOptions, v)
				continue
			}
			macros := utils.RemoveRepeated(re.FindAllString(strings.ReplaceAll(v, " ", ""), -1))
			if err := getMacroValue(macros, detectedMacroValue); err != nil {
				return fmt.Errorf("get detect value failed: %v", err)
			}
			calcResult, err := utils.Calculate(convertString(v, detectedMacroValue))
			if err != nil {
				return fmt.Errorf("'%v' calculate option err: %v", param.ParaName, err)
			}
			newOptions = append(newOptions, fmt.Sprintf("%v", int(calcResult)))
		}
		param.Options = newOptions
	}

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

	url := fmt.Sprintf("%v:%v/detect", config.KeenTune.DestIP, config.KeenTune.Group[0].Port)
	resp, err := http.RemoteCall("POST", url, requestMap)
	if err != nil {
		return fmt.Errorf("remote call err:%v", err)
	}

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

