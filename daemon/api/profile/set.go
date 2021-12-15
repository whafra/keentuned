package profile

import (
	com "keentune/daemon/api/common"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"keentune/daemon/common/utils/http"
	m "keentune/daemon/modules"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
)

type SetFlag struct {
	Name  string
}

// Set run profile set service
func (s *Service) Set(flag SetFlag, reply *string) error {
	defer func() {
		config.ProgramNeedExit <- true
		<-config.ServeFinish
		*reply = log.ClientLogMap[log.ProfSet]	
		log.ClearCliLog(log.ProfSet)
	}()

	go com.HeartbeatCheck()

	configFile, err := checkProfilePath(flag.Name)
	if err != nil{
		log.Errorf(log.ProfSet, "Check file [%v] err:%v", flag.Name, err)
		return nil
	}
	
	requestInfo, err := getConfigParamInfo(configFile)
	if err != nil {
		log.Errorf(log.ProfSet, "Get request info from specified file [%v] err:%v", flag.Name, err)
		return fmt.Errorf("Get request info from specified file [%v] err:%v", flag.Name, err)
	}

	if err := prepareBeforeSet(requestInfo); err != nil {
		log.Errorf(log.ProfSet, "Prepare before Set err:%v", err)
		return fmt.Errorf("Prepare before Set err:%v", err)
	}

	setResult, err := setConfiguration(requestInfo)
	if err != nil {
		log.Errorf(log.ProfSet, "Set failed:%v", err)
		return fmt.Errorf("Set failed:%v", err)
	}
	
	activeFile := m.GetProfileWorkPath("active.conf")
	if err := updateActiveFile(activeFile, []byte(file.GetPlainName(flag.Name))); err != nil {
		log.Errorf(log.ProfSet, "Update active file err:%v", err)
		return fmt.Errorf("Update active file err:%v", err)

	}

	log.Infof(log.ProfSet, "%v Set %v successfully: %v", utils.ColorString("green", "[OK]"), flag.Name, setResult)

	return nil
}

func checkProfilePath(name string) (string, error) {
	filePath := com.GetProfilePath(name)
	if filePath != "" {
		return filePath, nil
	}

	return "", fmt.Errorf("find the configuration file [%v] neither in[%v] nor in [%v]", name, fmt.Sprintf("%s/profile", config.KeenTune.Home), fmt.Sprintf("%s/profile", config.KeenTune.DumpHome))
}

func prepareBeforeSet(configInfo map[string]interface{}) error {	
	host := fmt.Sprintf("%s:%s", config.KeenTune.TargetIP, config.KeenTune.TargetPort)
	// step1. rollback the target machine
	if err := http.ResponseSuccess("POST", host + "/rollback", nil); err != nil {
		return fmt.Errorf("exec rollback failed, err:%v", err)
	}

	// step2. clear the active file
	fileName := m.GetProfileWorkPath("active.conf")
	if err := updateActiveFile(fileName, []byte{}); err != nil {
		return fmt.Errorf("update active file failed, err:%v", err)
	}

	// step3. backup the target machine
	if err := backup(host+"/backup", utils.Parse2Map("data", configInfo)); err != nil {
		return fmt.Errorf("exec backup failed, err:%v", err)
	}

	return nil
}

func getConfigParamInfo(configFile string) (map[string]interface{}, error) {
	resultMap, err := file.ConvertConfFileToJson(configFile)
	if err != nil {
		return nil, fmt.Errorf("convert file :%v err:%v", configFile, err)
	}

	respIP, err := utils.GetExternalIP()
	if err != nil {
		return nil, fmt.Errorf("run benchmark get real keentuned ip err: %v", err)
	}

	retRequst := map[string]interface{}{}	
	retRequst["data"] = resultMap
	retRequst["resp_ip"] = respIP
	retRequst["resp_port"] = config.KeenTune.Port

	return retRequst, nil
}

func setConfiguration(request interface{}) (string, error) {
	uri := fmt.Sprintf("%s:%s/configure", config.KeenTune.TargetIP, config.KeenTune.TargetPort)

	resp, err := http.RemoteCall("POST", uri, request)
	if err != nil {
		return "", err
	}

	setResult, err := parseSetResponse(resp)
	if err != nil {
		return "", err
	}

	return setResult, nil
}

func updateActiveFile(fileName string, info []byte) error {
	if err := ioutil.WriteFile(fileName, info, os.ModePerm); err != nil {
		return err
	}

	return nil
}

func backup(uri string, request interface{}) error {
	return http.ResponseSuccess("POST", uri, request)
}

func parseSetResponse(sucBytes []byte) (string, error) {
	applyMap, err := m.GetApplyResult(sucBytes)
	if err != nil {
		return "", err
	}

	var appliedInfo struct {
		Dtype   string      `json:"dtype"`
		Value   interface{} `json:"value"`
		Msg     string      `json:"msg"`
		Success bool        `json:"suc"`
	}

	var sucCount, failedCount int
	failedInfo := fmt.Sprintf("%s,%s;", "param name", "failed reason")

	for _, paramMaps := range applyMap {
		paramMap, ok := paramMaps.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("[parseSetResponse] assert param type [%v] to map failed", reflect.TypeOf(paramMaps))
		}

		for name, orgValue := range paramMap {
			orgParamMap, ok := orgValue.(map[string]interface{})
			if !ok {
				return "", fmt.Errorf("[parseSetResponse] assert param orgin value type[%v] to map failed", reflect.TypeOf(orgValue))
			}

			if err = utils.Map2Struct(orgParamMap, &appliedInfo); err != nil {
				return "", fmt.Errorf("[parseSetResponse] MapToStruct err:[%v]", err)
			}

			if appliedInfo.Success {
				sucCount++
				continue
			}

			failedCount++
			failedInfo += fmt.Sprintf("%s,%s;", name, appliedInfo.Msg)
		}

	}

	var setResult string

	if failedCount == 0 {
		setResult = fmt.Sprintf("total param %v, successed %v, failed %v.", sucCount, sucCount, failedCount)

		return setResult, nil
	}

	setResult = fmt.Sprintf("total param %v, successed %v, failed %v; the failed details displayed in the terminal.%s show table end.", sucCount+failedCount, sucCount, failedCount, failedInfo)

	return setResult, nil
}

