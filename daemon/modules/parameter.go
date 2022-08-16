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
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Parameter define a os parameter value scope, operating command and value
type Parameter struct {
	DomainName string        `json:"domain,omitempty"`
	ParaName   string        `json:"name,omitempty"`
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

const (
	defMarcoString = "#!([0-9A-Za-z_]+)#"
	recommendReg   = "^recommend.*"
)

// updateParameter update the partial param by the total param
func updateParameter(partial, total *Parameter) {
	if partial.Dtype == "" {
		partial.Dtype = total.Dtype
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
			calcResult, err := getExtremeValue(macros, detectedMacroValue, macroString)
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
			calcResult, err := getExtremeValue(macros, detectedMacroValue, v)
			if err != nil {
				return fmt.Errorf("'%v' calculate option err: %v", param.ParaName, err)
			}

			newOptions = append(newOptions, fmt.Sprintf("%v", int(calcResult)))
		}
		param.Options = newOptions
	}

	return nil
}

func getExtremeValue(macros []string, detectedMacroValue map[string]int, macroString string) (int64, error) {
	if len(macros) == 0 {
		return 0, fmt.Errorf("range type is '%v', but macros length is 0", macroString)
	}

	if err := getMacroValue(macros, detectedMacroValue); err != nil {
		return 0, fmt.Errorf("get detect value failed: %v", err)
	}

	express, symbol, compareValue := convertString(macroString, detectedMacroValue)

	calcResult, err := utils.Calculate(express)
	if err != nil || len(compareValue) == 0 {
		return calcResult, err
	}

	switch symbol {
	case "MAX":
		return int64(math.Max(float64(calcResult), compareValue[0])), nil
	case "MIN":
		return int64(math.Min(float64(calcResult), compareValue[0])), nil
	}

	return calcResult, nil
}

func convertString(macroString string, macroMap map[string]int) (string, string, []float64) {
	retStr := strings.ReplaceAll(macroString, " ", "")
	for name, value := range macroMap {
		retStr = strings.ReplaceAll(retStr, name, fmt.Sprint(value))
	}

	var symbol string
	if len(retStr) > 4 {
		switch strings.ToUpper(retStr)[0:4] {
		case "MAX(", "MAX[":
			symbol = "MAX"
		case "MIN(", "MIN[":
			symbol = "MIN"
		default:
			return retStr, "", nil
		}

		macroParts := strings.Split(retStr[4:len(retStr)-1], ",")
		express := ""
		var compareInt []float64
		for _, part := range macroParts {
			value, err := strconv.ParseFloat(part, 64)
			if err != nil {
				express = part
			} else {
				compareInt = append(compareInt, value)
			}
		}

		return express, symbol, compareInt
	}

	return retStr, "", nil
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

	url := fmt.Sprintf("%v:%v/detect", config.KeenTune.BenchGroup[0].DestIP, config.KeenTune.Group[0].Port)
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

// ConvertConfFileToJson convert conf file to json
func ConvertConfFileToJson(fileName string) (string, map[string]map[string]interface{}, error) {
	paramBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", nil, fmt.Errorf("read file err: %v", err)
	}

	if len(paramBytes) == 0 {
		return "", nil, fmt.Errorf("read file is empty")
	}

	var resultMap = make(map[string]map[string]interface{})
	var domainMap = make(map[string][]map[string]interface{})

	commonDomain := ""
	recommends := ""
	var tmpRecommendMap = make(map[string][]string)
	replacedStr := strings.ReplaceAll(string(paramBytes), "：", ":")
	for _, line := range strings.Split(replacedStr, "\n") {
		pureLine := strings.Replace(strings.TrimSpace(line), "=", ":", 1)
		if len(pureLine) == 0 {
			continue
		}

		if strings.HasPrefix(pureLine, "#") {
			continue
		}

		if strings.Contains(pureLine, "[") {
			commonDomain = strings.Trim(strings.Trim(strings.TrimSpace(line), "]"), "[")
			continue
		}

		recommend, param, err := readLine(pureLine)
		if err != nil {
			fmt.Printf("read line [%v] err:%v\n", line, err)
			continue
		}

		if len(recommend) != 0 {
			tmpRecommendMap[commonDomain] = append(tmpRecommendMap[commonDomain], recommend)
			continue
		}

		domainMap[commonDomain] = append(domainMap[commonDomain], param)
	}

	for key, value := range tmpRecommendMap {
		recommends += fmt.Sprintf("[%v]\n%v\n", key, strings.Join(value, ""))
	}

	if len(domainMap) == 0 {
		if recommends != "" {
			return recommends, nil, nil
		}

		return recommends, nil, fmt.Errorf("domain '%v' content is empty", commonDomain)
	}

	for domain, paramSlice := range domainMap {
		if len(paramSlice) == 0 {
			return recommends, nil, fmt.Errorf("domain '%v' content is empty", commonDomain)
		}

		var paramMap = make(map[string]interface{})
		for _, paramInfo := range paramSlice {
			name, ok := paramInfo["name"].(string)
			if !ok {
				fmt.Printf("parse name from [%v] failed\n", paramInfo)
				continue
			}
			delete(paramInfo, "name")
			paramMap[name] = paramInfo
		}
		resultMap[domain] = paramMap
	}

	return recommends, resultMap, nil
}

func readLine(line string) (string, map[string]interface{}, error) {
	var pureLine string
	re, _ := regexp.Compile(defMarcoString)
	if re != nil && re.MatchString(line) {
		pureLine = line
	} else {
		// remove comments
		parts := strings.Split(line, "#")
		if len(parts) <= 0 {
			return "", nil, fmt.Errorf("empty line")
		}

		pureLine = parts[0]
	}

	paramSlice := strings.Split(pureLine, ":")
	partLen := len(paramSlice)
	switch {
	case partLen <= 1:
		return "", nil, fmt.Errorf("param %v length %v is invalid, required: 2", paramSlice, len(paramSlice))
	case partLen == 2:
		return getParam(paramSlice)
	default:
		newSlice := []string{paramSlice[0]}
		newSlice = append(newSlice, strings.Join(paramSlice[1:], ":"))
		return getParam(newSlice)
	}
}

func getParam(paramSlice []string) (string, map[string]interface{}, error) {
	var recommend string
	paramName := strings.TrimSpace(paramSlice[0])
	valueStr := strings.ReplaceAll(strings.TrimSpace(paramSlice[1]), "\"", "")
	matched, _ := regexp.MatchString(recommendReg, strings.ToLower(valueStr))
	if matched {
		recommend = fmt.Sprintf("\t%v: %v\n", paramName, strings.TrimPrefix(valueStr, "recommend:"))
		return recommend, nil, nil
	}

	re, _ := regexp.Compile(defMarcoString)
	if re != nil && re.MatchString(valueStr) {
		return detectConfValue(re, valueStr, paramName)
	}

	var param map[string]interface{}
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		param = map[string]interface{}{
			"value": valueStr,
			"dtype": "string",
			"name":  paramName,
		}
		return "", param, nil
	}

	param = map[string]interface{}{
		"value": value,
		"dtype": "int",
		"name":  paramName,
	}
	return "", param, nil
}

func detectConfValue(re *regexp.Regexp, valueStr string, paramName string) (string, map[string]interface{}, error) {
	macros := utils.RemoveRepeated(re.FindAllString(strings.ReplaceAll(valueStr, " ", ""), -1))
	detectedMacroValue := make(map[string]int)
	value, err := getExtremeValue(macros, detectedMacroValue, valueStr)
	if err != nil {
		return "", nil, fmt.Errorf("detect '%v'err: %v", paramName, err)
	}

	param := map[string]interface{}{
		"value": value,
		"dtype": "int",
		"name":  paramName,
	}

	return "", param, nil
}

