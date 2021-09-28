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
		log.Errorf(log.ProfSet, "check file [%v] err:%v", flag.Name, err)
		return err
	}
	
	requestInfo, err := getConfigParamInfo(configFile)
	if err != nil {
		log.Errorf(log.ProfSet, "get request info from specified file [%v] err:%v", flag.Name, err)
		return err
	}

	if err := prepareBeforeSet(requestInfo); err != nil {
		log.Errorf(log.ProfSet, "prepare before Set err:%v", err)
		return err
	}

	log.Infof(log.ProfSet, "\nStep1. [Set profile] get ready.")

	if err := setConfiguration(requestInfo); err != nil {
		log.Errorf(log.ProfSet, "failed to exec [profile set], err:%v", err)
		return nil
	}

	log.Infof(log.ProfSet, "\nStep2. [Set profile] set configuration successfully.")

	activeFile := m.GetProfileWorkPath("active.conf")
	if err := updateActiveFile(activeFile, []byte(flag.Name)); err != nil {
		log.Errorf(log.ProfSet, "update active file err:%v", err)
		return nil
	}

	log.Infof(log.ProfSet, "\nStep3. [Set profile] update active configure file successfully, and the process finish.")
	return nil
}

func checkProfilePath(name string) (string, error) {
	dumpProfPath :=m.GetProfileWorkPath(name)
	if file.IsPathExist(dumpProfPath) {
		return dumpProfPath, nil
	}

	demoProfPath :=fmt.Sprintf("%s/profile/%s", config.KeenTune.Home, name)
	if file.IsPathExist(demoProfPath) {
		return demoProfPath, nil
	}

	return "", fmt.Errorf("find the configuration file [%v] neither in[] nor in []", name, fmt.Sprintf("%s/profile", config.KeenTune.Home), fmt.Sprintf("%s/profile", config.KeenTune.DumpHome))
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

func setConfiguration(request interface{}) error {
	uri := fmt.Sprintf("%s:%s/configure", config.KeenTune.TargetIP, config.KeenTune.TargetPort)

	resp, err := http.RemoteCall("POST", uri, request)
	if err != nil {
		return err
	}

	if err = parseSetResponse(resp); err != nil {
		return err
	}

	return nil
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

func parseSetResponse(sucBytes []byte) error {
	applyMap, err := m.GetApplyResult(sucBytes)
	if err != nil {
		return err
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
			return fmt.Errorf("[parseSetResponse] assert param type [%v] to map failed", reflect.TypeOf(paramMaps))
		}

		for name, orgValue := range paramMap {
			orgParamMap, ok := orgValue.(map[string]interface{})
			if !ok {
				return fmt.Errorf("[parseSetResponse] assert param orgin value type[%v] to map failed", reflect.TypeOf(orgValue))
			}

			if err = utils.Map2Struct(orgParamMap, &appliedInfo); err != nil {
				return fmt.Errorf("[parseSetResponse] MapToStruct err:[%v]", err)
			}

			if appliedInfo.Success {
				sucCount++
				continue
			}

			failedCount++
			failedInfo += fmt.Sprintf("%s,%s;", name, appliedInfo.Msg)
		}

	}

	if failedCount == 0 {
		log.Infof(log.ProfSet, "\tSet result: total param counts: %v; successed: %v; failed: %v.", sucCount+failedCount, sucCount, failedCount)

		return nil
	}

	log.Infof(log.ProfSet, "\tSet result: total param counts: %v; successed: %v; failed: %v; the failed details displayed in the terminal.", sucCount+failedCount, sucCount, failedCount)
	
	log.Infof(log.ProfSet, "%s show table end.", failedInfo)
	return nil
}
