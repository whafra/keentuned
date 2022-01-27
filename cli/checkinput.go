package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	com "keentune/daemon/api/common"
	"keentune/daemon/api/param"
	"keentune/daemon/common/file"
	"keentune/daemon/common/utils"
	m "keentune/daemon/modules"
	"regexp"
	"strings"
)

func checkTuningFlags(cmdName string, flag *TuneFlag) error {
	var err error
	jobFlag := "--data"
	if cmdName == "tune" {
		jobFlag = "--job"
	}

	if err = checkJob(cmdName, flag.Name); err != nil {
		return fmt.Errorf("%v %v", jobFlag, err)
	}

	if flag.Round <= 0 {
		return fmt.Errorf("--iteration must be positive integer, input: %v", flag.Round)
	}

	if flag.ParamMap, err = checkParamConf(flag.ParamConf); err != nil {
		return fmt.Errorf("--param %v", err)
	}

	if err = checkBenchConf(&flag.BenchConf); err != nil {
		return fmt.Errorf("--bench %v", err)
	}

	return nil
}

func checkBenchConf(conf *string) error {
	if !strings.HasSuffix(*conf, ".json") {
		return fmt.Errorf("bench file suffix is not json")
	}

	benchConf := param.GetBenchJsonPath(*conf)
	if !file.IsPathExist(benchConf) {
		return fmt.Errorf("bench file [%v] does not exist", *conf)
	}

	*conf = benchConf

	reqData, err := ioutil.ReadFile(*conf)
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

func checkParamConf(confs []string) (string, error) {
	var mergedParam = make(map[string]map[string]interface{})
	if len(confs) == 0 {
		return "", fmt.Errorf("param file suffix is not json, param name is needed")
	}

	for _, conf := range confs {
		fileName := strings.Trim(conf, " ")
		if !strings.HasSuffix(fileName, ".json") {
			return "", fmt.Errorf("param file suffix is not json")
		}

		paramConf := com.GetAbsolutePath(fileName, "parameter", ".json", "_best.json")
		if !file.IsPathExist(paramConf) {
			return "", fmt.Errorf("param file [%v] does not exist", fileName)
		}

		userParamMap, err := readFile(paramConf)
		if err != nil {
			return "", err
		}

		err = readParams(userParamMap, mergedParam)
		if err != nil {
			return "", fmt.Errorf("check %v file: %v", fileName, err)
		}
	}

	paramBytes, err := json.Marshal(mergedParam)
	if err != nil {
		return "", err
	}

	return string(paramBytes), nil
}

func readFile(fileName string) (map[string]map[string]interface{}, error) {
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("read [%v] file:%v\n", fileName, err)
	}

	if len(bytes) == 0 {
		return nil, fmt.Errorf("file is empty")
	}

	var retMap map[string]map[string]interface{}
	err = json.Unmarshal(bytes, &retMap)
	if err == nil && len(retMap) != 0 {
		return retMap, nil
	}

	var paramMap map[string]interface{}
	err = json.Unmarshal(bytes, &paramMap)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal err:%v\n", err)
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

func readParams(userParamMap, mergedParam map[string]map[string]interface{}) error {
	var err error
	for domainName, domainMap := range userParamMap {
		_, ok := mergedParam[domainName]
		if !ok {
			mergedParam[domainName] = make(map[string]interface{})
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

			if _, ok = mergedParam[domainName][name]; !ok {
				mergedParam[domainName][name] = paramMap
			}
		}
	}

	return nil
}

func checkParam(name string, paramMap map[string]interface{}) error {
	var param m.Parameter
	err := utils.Map2Struct(paramMap, &param)
	if err != nil {
		return fmt.Errorf("map to struct err:%v", err)
	}

	param.ParaName = name
	// check data type
	if !isDataTypeOK(param.Dtype) {
		return fmt.Errorf("param %v data type must be one of int, float, string or bool. find: %v", param.ParaName, param.Dtype)
	}

	// check range length=2
	if len(param.Scope) == 2 {
		range1, ok1 := param.Scope[0].(float64)
		range2, ok2 := param.Scope[1].(float64)
		if ok1 && ok2 {
			if range2 <= range1 {
				return fmt.Errorf("param %v range[1] must be larger than range[0]", param.ParaName)
			}
		}
	}

	return checkUniqueField(param)
}

func checkUniqueField(param m.Parameter) error {
	if len(param.Scope) == 0 && len(param.Sequence) == 0 && len(param.Options) == 0 {
		return fmt.Errorf("param %v field range, options and sequence, only one of them can exist", param.ParaName)
	}
	if len(param.Scope) > 0 && len(param.Sequence) > 0 {
		return fmt.Errorf("param %v range and sequence, only one of them can exist", param.ParaName)
	}

	if len(param.Scope) > 0 && len(param.Options) > 0 {
		return fmt.Errorf("param %v range and options, only one of them can exist", param.ParaName)
	}

	if len(param.Sequence) > 0 && len(param.Options) > 0 {
		fmt.Printf("%v param %vsequence and options, only one of them can exist\n", ColorString("yellow", "[Warning]"), param.ParaName)
	}

	if (param.Dtype == "string" || param.Dtype == "str") && param.Step > 0.0 {
		return fmt.Errorf("param %v 'step' field is not supported for data type %v", param.ParaName, param.Dtype)
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

func checkJob(cmd, name string) error {
	re := regexp.MustCompile("[^A-Za-z0-9_]+")
	re.FindAll([]byte(name), -1)

	var result string
	for _, value := range re.FindAllString(name, -1) {
		for _, v := range value {
			result += fmt.Sprintf("%q ", v)
		}
	}

	if len(result) != 0 {
		return fmt.Errorf("find unexpected characters %v. Only \"a-z\", \"A-Z\", \"0-9\" and \"_\" are supported", result)
	}

	if cmd == "tune" && isTuneNameRepeat(name) {
		return fmt.Errorf("the specified name [%v] already exists. Run [keentune param delete --job %v] or specify a new name and try again", name, name)
	}

	return nil
}

func isTuneNameRepeat(name string) bool {
	tuneList, err := file.WalkFilePath(m.GetTuningWorkPath("")+"/", "", true, "/generate/")
	if err != nil {
		return false
	}

	for _, has := range tuneList {
		if has == name {
			return true
		}
	}

	return false
}
