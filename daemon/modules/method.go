package modules

import (
	"encoding/json"
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/utils/http"
	"regexp"
	"strings"
)

const (
	defDoubleFuncReg       = "\\$\\{f:(.*):\\$\\{f:(.*):(.*):(.*)\\}\\}"
	defFuncWithAssertReg   = "\\$\\{f:(.*):(.*):(.*):(.*)\\}"
	defFuncWithOneArgReg   = "\\$\\{f:(.*):(.*)\\}"
	defFuncWithFourArgsReg = "\\$\\{f:(.*):(.*):(.*):(.*):(.*)\\}"
)

const (
	noBalanceCores      = "no_balance_cores"
	isolatedCores       = "isolated_cores"
	isolatedCoresAssert = "isolated_cores_assert_check"
)

var cpuCoresSpecValue = map[string]string{
	"no_balance_cores":            "2-3",
	"isolated_cores":              "5",
	"isolated_cores_assert_check": "\\2-3",
}

type methodReq struct {
	Name string        `json:"method_name"`
	Args []interface{} `json:"method_args"`
}

type methodResp struct {
	Suc    bool   `json:"suc"`
	Result string `json:"res"`
}

func getMethodReqByNames(data []string) []methodReq {
	var retReqMethod []methodReq
	for _, methodName := range data {
		retReqMethod = append(retReqMethod, methodReq{
			Name: methodName,
			Args: []interface{}{},
		})
	}

	return retReqMethod
}

func getMethodReqByArg(data map[string]interface{}) (map[string]string, []methodReq) {
	var retReqMethod []methodReq
	var varNames = make(map[string]string)
	for name, arg := range data {
		req := arg.(methodReq)
		retReqMethod = append(retReqMethod, arg.(methodReq))
		varNames[name] = req.Name
	}

	return varNames, retReqMethod
}

func matchString(pattern, s string) bool {
	res, err := regexp.MatchString(pattern, s)
	if err != nil {
		return false
	}

	return res
}

func getVariableReq(line string, varMap map[string]interface{}) {
	parts := strings.Split(line, ":")
	if len(parts) < 2 {
		return
	}
	varName := strings.TrimSpace(parts[0])
	varRegexStr := strings.TrimSpace(strings.Join(parts[1:], ":"))

	switch {
	case matchString(defDoubleFuncReg, varRegexStr):
		req := getDoubleFuncMethodReq(varRegexStr)
		varMap[varName] = req
	case matchString(defFuncWithFourArgsReg, varRegexStr):
		// todo
	case matchString(defFuncWithAssertReg, varRegexStr):
		// todo
	case matchString(defFuncWithOneArgReg, varRegexStr):
		req := getFuncWithOneArgMethodReq(varRegexStr, varMap)
		varMap[varName] = req
	default:
		return
	}
}

func getFuncWithOneArgMethodReq(origin string, varMap map[string]interface{}) methodReq {
	reg := regexp.MustCompile(defFuncWithOneArgReg)
	replaced := reg.ReplaceAllString(origin, "$1#$ $2")
	args := strings.Split(replaced, "#$ ")
	if len(args) != 2 {
		return methodReq{
			Name: args[0],
			Args: []interface{}{},
		}
	}

	if matchString("\\$\\{(.*)\\}", args[1]) {
		varName := strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(args[1]), "${"), "}")
		specValue, find := cpuCoresSpecValue[varName]
		var arg interface{}
		if find {
			arg = specValue
		} else {
			arg = varMap[varName]
		}

		return methodReq{
			Name: args[0],
			Args: []interface{}{arg},
		}
	}

	return methodReq{
		Name: args[0],
		Args: []interface{}{args[1]},
	}
}

func getDoubleFuncMethodReq(origin string) methodReq {
	reg := regexp.MustCompile(defDoubleFuncReg)
	replaced := reg.ReplaceAllString(origin, "$1#$ $2#$ $3#$ $4")
	args := strings.Split(replaced, "#$ ")
	if len(args) != 4 {
		return methodReq{
			Name: args[0],
			Args: []interface{}{},
		}
	}

	innerReq := methodReq{
		Name: args[1],
		Args: []interface{}{args[2], args[3]},
	}

	return methodReq{
		Name: args[0],
		Args: []interface{}{innerReq},
	}
}

func requestAllVariables(destMap map[string]string, reqMap map[string]interface{}) error {
	varNames, req := getMethodReqByArg(reqMap)
	url := fmt.Sprintf("%v:%v/method", config.KeenTune.BenchGroup[0].DestIP, config.KeenTune.Group[0].Port)
	resp, err := http.RemoteCall("POST", url, req)
	if err != nil {
		return fmt.Errorf("remote call err:%v", err)
	}

	var respMap map[string]methodResp
	err = json.Unmarshal(resp, &respMap)
	if err != nil {
		return fmt.Errorf("unmarshal detect response err:%v", err)
	}

	for varName, funcName := range varNames {
		result, ok := respMap[funcName]
		if !ok {
			return fmt.Errorf("get response %v err:%v", funcName, err)
		}
		if !result.Suc {
			return fmt.Errorf("get response %v failed:%v", funcName, result.Result)
		}
		destMap[varName] = result.Result
	}

	return nil
}

