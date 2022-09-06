package config

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"keentune/daemon/common/file"
	"keentune/daemon/common/utils"
	"strings"
)

const hexTable = "0123456789abcdef"

var KeenTuneConfMD5 string

type Parameter struct {
	Domain   string        `json:"domain"`
	Name     string        `json:"name"`
	Scope    []interface{} `json:"range,omitempty"`
	Options  []string      `json:"options,omitempty"`
	Sequence []interface{} `json:"sequence,omitempty"`
	Dtype    string        `json:"dtype"`
	Value    interface{}   `json:"value,omitempty"`
	Step     int           `json:"step,omitempty"`
	Weight   float32       `json:"weight,omitempty"`
}

func checkBenchConf(conf *string) error {
	if !strings.HasSuffix(*conf, ".json") {
		return fmt.Errorf("bench file suffix is not json")
	}

	benchConf := GetBenchJsonPath(*conf)
	if !file.IsPathExist(benchConf) {
		return fmt.Errorf("bench file [%v] does not exist", *conf)
	}

	reqData, err := ioutil.ReadFile(benchConf)
	if err != nil {
		return fmt.Errorf("read bench conf file err: %v", err)
	}

	var bench map[string]interface{}

	if err = json.Unmarshal(reqData, &bench); err != nil {
		return fmt.Errorf("unmarshal bench conf file err: %v", err)
	}

	benchInterface, ok := bench["benchmark"]
	benchList, ok := benchInterface.([]interface{})
	if len(benchList) == 0 || !ok {
		return fmt.Errorf("benchmark field doesn't exist")
	}

	for i, benchMap := range benchList {
		value, ok := benchMap.(map[string]interface{})
		if !ok {
			return fmt.Errorf("benchmark type is not struct")
		}

		if err = parse2String(value, "benchmark_cmd"); err != nil {
			return fmt.Errorf("%vth bench benchmark_cmd %v", i+1, err)
		}

		if err = parse2String(value, "local_script_path"); err != nil {
			return fmt.Errorf("%vth bench local_script_path %v", i+1, err)
		}

		if err = checkItem(value); err != nil {
			return fmt.Errorf("%vth bench items %v", i+1, err)
		}
	}

	return nil
}

func checkItem(value map[string]interface{}) error {
	itemMap, ok := value["items"]
	if !ok {
		return fmt.Errorf("field doesn't exist")
	}

	items, ok := itemMap.(map[string]interface{})
	if !ok {
		return fmt.Errorf("field is not struct")
	}

	if len(items) == 0 {
		return fmt.Errorf("is null")
	}

	var err error
	var zeroCount int
	for key, item := range items {
		itemInfo, ok := item.(map[string]interface{})
		if !ok {
			return fmt.Errorf("[%v] is not struct", key)
		}

		if err = parse2Bool(itemInfo, "negative"); err != nil {
			return fmt.Errorf("[%v] negative %v", key, err)
		}

		if err = parse2Bool(itemInfo, "strict"); err != nil {
			return fmt.Errorf("[%v] strict %v", key, err)
		}

		weight, err := parse2Float(itemInfo, "weight")
		if err != nil {
			return fmt.Errorf("[%v] weight %v", key, err)
		}

		if weight < 0.0 {
			return fmt.Errorf("[%v] weight is less than 0.0", key)
		}
		if weight == 0.0 {
			zeroCount++
		}
	}

	if zeroCount == len(items) {
		return fmt.Errorf("at least one weight must be greater than 0.0")
	}

	return nil
}

func parse2String(origin map[string]interface{}, key string) error {
	value, ok := origin[key]
	if !ok {
		return fmt.Errorf("field doesn't exist")
	}

	valueStr, ok := value.(string)
	if !ok {
		return fmt.Errorf("field is not string type")
	}

	if strings.Trim(valueStr, " ") == "" {
		return fmt.Errorf("field is empty")
	}

	return nil
}

func parse2Bool(origin map[string]interface{}, key string) error {
	value, ok := origin[key]
	if !ok {
		return fmt.Errorf("field doesn't exist")
	}

	_, ok = value.(bool)
	if !ok {
		return fmt.Errorf("field is not boolen type")
	}

	return nil
}

func parse2Float(origin map[string]interface{}, key string) (float32, error) {
	value, ok := origin[key]
	if !ok {
		return 0, fmt.Errorf("field doesn't exist")
	}

	val, ok := value.(float64)
	if !ok {
		return 0, fmt.Errorf("field is not float or int type")
	}

	return float32(val), nil
}

func checkParamConf(confs []string) (map[string]string, []DBLMap, error) {
	if len(confs) == 0 {
		return nil, nil, fmt.Errorf("param file suffix is not json, param name is needed")
	}

	var domains = make(map[string]string)
	var mergedParam = make([]DBLMap, PRILevel)
	for _, conf := range confs {
		fileName := strings.Trim(conf, " ")
		if !strings.HasSuffix(fileName, ".json") {
			return nil, nil, fmt.Errorf("param file suffix is not json")
		}

		paramConf := GetAbsolutePath(fileName, "parameter", ".json", "_best.json")
		if !file.IsPathExist(paramConf) {
			return nil, nil, fmt.Errorf("param file [%v] does not exist", fileName)
		}

		userParamMap, err := readFile(paramConf)
		if err != nil {
			return nil, nil, err
		}

		err = readParams(domains, userParamMap, mergedParam)
		if err != nil {
			return nil, nil, fmt.Errorf("check %v file: %v", fileName, err)
		}
	}

	return domains, mergedParam, nil
}

func readFile(fileName string) (DBLMap, error) {
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("read [%v] file:%v\n", fileName, err)
	}

	if len(bytes) == 0 {
		return nil, fmt.Errorf("file is empty")
	}

	var retMap DBLMap
	err = json.Unmarshal(bytes, &retMap)
	if err == nil && len(retMap) != 0 {
		return retMap, nil
	}

	var paramMap map[string]interface{}
	err = json.Unmarshal(bytes, &paramMap)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal err: %v", err)
	}

	var domains []string
	for domain, _ := range paramMap {
		domains = append(domains, domain)
	}

	if len(domains) == 0 {
		return nil, fmt.Errorf("assert domain does not exist")
	}

	return nil, fmt.Errorf("assert domain %v value is not matched, such as: {\"domain\":{\"param1\":{\"dtype\":\"string\",\"options\":[\"0\",\"1\"]}}}", domains[0])
}

func readParams(domains map[string]string, userParamMap DBLMap, mergedParam []DBLMap) error {
	var err error
	for domainName, domainMap := range userParamMap {
		priID, ok := PriorityList[domainName]
		if !ok {
			PriorityList[domainName] = 1
			priID = 1
		}

		if mergedParam[priID] == nil {
			mergedParam[priID] = make(DBLMap)
		}
		_, ok = mergedParam[priID][domainName]
		if !ok {
			mergedParam[priID][domainName] = make(map[string]interface{})
		}

		for name, paramValue := range domainMap {
			paramMap, ok := paramValue.(map[string]interface{})
			if !ok {
				return fmt.Errorf("parse param %v value [%+v] type to map failed", name, paramValue)
			}

			// check step
			stepInterface, ok := paramMap["step"]
			if ok {
				step, ok := stepInterface.(float64)
				if ok && step <= 0.0 {
					return fmt.Errorf("param %v step must be larger than 0, find: %v", name, step)
				}
			}

			if err = checkParam(name, paramMap); err != nil {
				return err
			}

			if _, ok = mergedParam[priID][domainName][name]; !ok {
				mergedParam[priID][domainName][name] = paramMap
			}
		}
		domains[domainName] = domainName
	}

	return nil
}

func checkParam(name string, paramMap map[string]interface{}) error {
	var param Parameter
	err := utils.Map2Struct(paramMap, &param)
	if err != nil {
		return fmt.Errorf("map to struct err:%v", err)
	}

	param.Name = name
	// check data type
	if !isDataTypeOK(param.Dtype) {
		return fmt.Errorf("param %v data type must be one of int, float, string or bool. find: %v", param.Name, param.Dtype)
	}

	// check range length=2
	if len(param.Scope) == 2 {
		range1, ok1 := param.Scope[0].(float64)
		range2, ok2 := param.Scope[1].(float64)
		if ok1 && ok2 {
			if range2 <= range1 {
				return fmt.Errorf("param %v range[1] must be larger than range[0]", param.Name)
			}
		}
	}

	return checkUniqueField(param)
}

func checkUniqueField(param Parameter) error {
	if len(param.Scope) == 0 && len(param.Sequence) == 0 && len(param.Options) == 0 {
		return fmt.Errorf("param %v field range, options and sequence, only one of them can exist", param.Name)
	}
	if len(param.Scope) > 0 && len(param.Sequence) > 0 {
		return fmt.Errorf("param %v range and sequence, only one of them can exist", param.Name)
	}

	if len(param.Scope) > 0 && len(param.Options) > 0 {
		return fmt.Errorf("param %v range and options, only one of them can exist", param.Name)
	}

	if len(param.Sequence) > 0 && len(param.Options) > 0 {
		fmt.Printf("%v param %vsequence and options, only one of them can exist\n", utils.ColorString("yellow", "[Warning]"), param.Name)
	}

	if (param.Dtype == "string" || param.Dtype == "str") && param.Step > 0.0 {
		return fmt.Errorf("param %v 'step' field is not supported for data type %v", param.Name, param.Dtype)
	}

	return nil
}

func isDataTypeOK(dtype string) bool {
	switch strings.Trim(dtype, " ") {
	case "int":
		return true
	case "float":
		return true
	case "string", "str":
		return true
	case "bool":
		return true
	default:
		return false
	}
}

// GetPriorityParams get param array by Priority
func GetPriorityParams(userParamMap DBLMap) ([]DBLMap, error) {
	var mergedParam = make([]DBLMap, PRILevel)
	for domainName, domainMap := range userParamMap {
		priID, ok := PriorityList[domainName]
		if !ok {
			PriorityList[domainName] = 1
			priID = 1
		}

		if mergedParam[priID] == nil {
			mergedParam[priID] = make(DBLMap)
		}
		_, ok = mergedParam[priID][domainName]
		if !ok {
			mergedParam[priID][domainName] = make(map[string]interface{})
		}

		for name, paramValue := range domainMap {
			paramMap, ok := paramValue.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("parse param %v value [%+v] type to map failed", name, paramValue)
			}

			if _, ok = mergedParam[priID][domainName][name]; !ok {
				mergedParam[priID][domainName][name] = paramMap
			}
		}
	}
	return mergedParam, nil
}

// GetKeenTuneConfFileMD5 get md5 of keentuned.conf file
func GetKeenTuneConfFileMD5() string {
	infos, _ := ioutil.ReadFile(keentuneConfigFile)
	src := md5.Sum(infos)
	var dst = make([]byte, 32)
	j := 0
	for _, v := range src {
		dst[j] = hexTable[v>>4]
		dst[j+1] = hexTable[v&0x0f]
		j += 2
	}

	return string(dst)
}

