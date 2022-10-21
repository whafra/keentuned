package modules

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
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

// DetectResult detect response
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

	detectFile := fmt.Sprintf("%v/detect/detect.json", config.KeenTune.Home)
	bytes, err := ioutil.ReadFile(detectFile)
	if err != nil {
		return fmt.Errorf("read detect json file err:%v", err)
	}

	var macroExpMap, stockMacroMap map[string]string
	err = json.Unmarshal(bytes, &macroExpMap)
	if err != nil {
		return fmt.Errorf("unmarshal detect json file err:%v", err)
	}

	stockMacroMap = make(map[string]string)
	for macro, expression := range macroExpMap {
		stockMacroMap[strings.ToLower(macro)] = expression
	}

	var macroNames []string
	var macroMap = make(map[string]string)
	for _, macro := range macros {
		if _, ok := detectedMacroValue[macro]; ok {
			continue
		}

		name := strings.TrimSuffix(strings.TrimPrefix(macro, "#!"), "#")
		lowerName := strings.ToLower(name)

		macroNames = append(macroNames, name)

		macroCmdExp, find := stockMacroMap[lowerName]
		if !find {
			return fmt.Errorf("detect can't find matched macro: %v", macro)
		}

		macroMap[lowerName] = macroCmdExp
	}

	if len(macroMap) == 0 {
		return nil
	}

	return detect(macroMap, macroNames, detectedMacroValue)
}

// ConvertConfFileToJson convert conf file to json
func ConvertConfFileToJson(fileName string) (ABNLResult, map[string]map[string]interface{}, error) {
	var abnormal = ABNLResult{}
	paramBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return abnormal, nil, fmt.Errorf("read file err: %v", err)
	}

	if len(paramBytes) == 0 {
		return abnormal, nil, fmt.Errorf("read file is empty")
	}

	replacedStr := strings.ReplaceAll(string(paramBytes), "：", ":")
	commonDomain, recommendMap, domainMap := parseConfStrToMapSlice(replacedStr, fileName, &abnormal)

	for key, value := range recommendMap {
		abnormal.Recommend += fmt.Sprintf("\t[%v]\n%v", key, strings.Join(value, ""))
	}

	if len(domainMap) == 0 {
		if abnormal.Recommend != "" {
			return abnormal, nil, nil
		}

		return abnormal, nil, fmt.Errorf("domain '%v' content is empty", commonDomain)
	}

	return changeMapSliceToDBLMap(domainMap, abnormal)
}

func changeMapSliceToDBLMap(domainMap map[string][]map[string]interface{}, abnormal ABNLResult) (ABNLResult, map[string]map[string]interface{}, error) {
	var resultMap = make(map[string]map[string]interface{})
	for domain, paramSlice := range domainMap {
		if len(paramSlice) == 0 {
			return abnormal, nil, fmt.Errorf("domain '%v' content is empty", domain)
		}

		var paramMap = make(map[string]interface{})
		for _, paramInfo := range paramSlice {
			name, ok := paramInfo["name"].(string)
			if !ok {
				abnormal.Warning += fmt.Sprintf("%v name does not exist", paramInfo)
				continue
			}
			delete(paramInfo, "name")
			paramMap[name] = paramInfo
		}
		resultMap[domain] = paramMap
	}
	return abnormal, resultMap, nil
}

func parseConfStrToMapSlice(replacedStr, fileName string, abnormal *ABNLResult) (string, map[string][]string, map[string][]map[string]interface{}) {
	var deleteDomains []string
	var recommendMap = make(map[string][]string)
	var domainMap = make(map[string][]map[string]interface{})
	var commonDomain string
	for _, line := range strings.Split(replacedStr, "\n") {
		pureLine := strings.TrimSpace(replaceEqualSign(line))
		if len(pureLine) == 0 {
			continue
		}

		if strings.HasPrefix(pureLine, "#") {
			continue
		}

		if strings.Contains(pureLine, "[") && !strings.Contains(pureLine, "∈") {
			commonDomain = strings.Trim(strings.Trim(strings.TrimSpace(line), "]"), "[")
			continue
		}

		recommend, condition, param, err := readLine(pureLine)
		if len(condition) != 0 {
			deleteDomains = append(deleteDomains, commonDomain)
			if commonDomain == myConfDomain {
				notMetInfo := fmt.Sprintf(detectENVNotMetFmt, commonDomain, myConfCondition, file.GetPlainName(fileName))
				abnormal.Warning += fmt.Sprintf("%v%v", notMetInfo, multiRecordSeparator)
				continue
			}

			notMetInfo := fmt.Sprintf(detectENVNotMetFmt, commonDomain, condition, file.GetPlainName(fileName))
			abnormal.Warning += fmt.Sprintf("%v%v", notMetInfo, multiRecordSeparator)

			continue
		}

		if err != nil {
			abnormal.Warning += fmt.Sprintf("content '%v' abnormal%v", pureLine, multiRecordSeparator)
			continue
		}

		if len(recommend) != 0 {
			recommendMap[commonDomain] = append(recommendMap[commonDomain], recommend)
			continue
		}

		// when condition is empty, param maybe null
		if param == nil {
			continue
		}

		domainMap[commonDomain] = append(domainMap[commonDomain], param)
	}

	if len(deleteDomains) > 0 {
		for _, domain := range deleteDomains {
			delete(domainMap, domain)
		}
	}

	return commonDomain, recommendMap, domainMap
}

func readLine(line string) (string, string, map[string]interface{}, error) {
	paramSlice := strings.Split(line, ":")
	partLen := len(paramSlice)
	switch {
	case partLen <= 1:
		return "", "", nil, fmt.Errorf("param %v length %v is invalid, required: 2", paramSlice, len(paramSlice))
	case partLen == 2:
		return getParam(paramSlice)
	default:
		newSlice := []string{paramSlice[0]}
		newSlice = append(newSlice, strings.Join(paramSlice[1:], ":"))
		return getParam(newSlice)
	}
}

func getParam(paramSlice []string) (string, string, map[string]interface{}, error) {
	var recommend string
	paramName := strings.TrimSpace(paramSlice[0])
	valueStr := strings.ReplaceAll(strings.TrimSpace(paramSlice[1]), "\"", "")

	matched, _ := regexp.MatchString(recommendReg, strings.ToLower(valueStr))
	if matched {
		forceWrapLine := strings.Replace(valueStr, ". Please", ".\n\t\t\tPlease", 1)
		recommend = fmt.Sprintf("\t\t%v: %v\n", paramName, strings.TrimPrefix(forceWrapLine, "recommend:"))
		return recommend, "", nil, nil
	}

	re, _ := regexp.Compile(defMarcoString)
	if re != nil && re.MatchString(valueStr) {
		expression, param, err := detectConfValue(re, valueStr, paramName)
		// replace expression to real condition when expression is not empty
		if expression != "" {
			return "", valueStr, nil, nil
		}

		return "", "", param, err
	}

	var param map[string]interface{}
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		param = map[string]interface{}{
			"value": valueStr,
			"dtype": "string",
			"name":  paramName,
		}
		return "", "", param, nil
	}

	param = map[string]interface{}{
		"value": value,
		"dtype": "int",
		"name":  paramName,
	}
	return "", "", param, nil
}

func replaceEqualSign(origin string) string {
	equalIdx := strings.Index(origin, "=")
	colonIdx := strings.Index(origin, ":")
	// First, '=' exists; if ':' not exist or '=' before ':', replace '=' by ':'
	if equalIdx > 0 && (colonIdx < 0 || equalIdx < colonIdx) {
		return strings.Replace(origin, "=", ":", 1)
	}

	return origin
}

