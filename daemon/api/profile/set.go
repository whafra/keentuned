package profile

import (
	"fmt"
	"io/ioutil"
	com "keentune/daemon/api/common"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"keentune/daemon/common/utils/http"
	m "keentune/daemon/modules"
	"os"
	"strings"
	"sync"
)

type SetFlag struct {
	Name string
}
type Result struct {
	Info    string
	Success bool
}

// Set run profile set service
func (s *Service) Set(flag SetFlag, reply *string) error {
	if com.IsApplying() {
		return fmt.Errorf("operation does not support, job %v is running", com.GetRunningTask())
	}

	defer func() {
		config.ProgramNeedExit <- true
		<-config.ServeFinish
		*reply = log.ClientLogMap[log.ProfSet]
		log.ClearCliLog(log.ProfSet)
	}()

	if err := com.HeartbeatCheck(); err != nil {
		log.Errorf(log.ProfSet, "Check %v", err)
		return fmt.Errorf("Check %v", err)
	}

	configFile, err := checkProfilePath(flag.Name)
	if err != nil {
		log.Errorf(log.ProfSet, "Check file [%v] err:%v", flag.Name, err)
		return fmt.Errorf("Check file [%v] err:%v", flag.Name, err)
	}

	requestInfo, err := getConfigParamInfo(configFile)
	if err != nil {
		log.Errorf(log.ProfSet, "Get request info from specified file [%v] err:%v", flag.Name, err)
		return fmt.Errorf("Get request info from specified file [%v] err:%v", flag.Name, err)
	}

	if err := prepareBeforeSet(requestInfo); err != nil {
		log.Errorf(log.ProfSet, "Prepare for Set err:%v", err)
		return fmt.Errorf("Prepare for Set err:%v", err)
	}

	sucInfos, failedInfo, err := setConfiguration(requestInfo)
	if err != nil {
		log.Errorf(log.ProfSet, "Set failed:%v, details:%v", err, failedInfo)
		return fmt.Errorf("Set failed:%v, details:%v", err, failedInfo)
	}

	activeFile := config.GetProfileWorkPath("active.conf")
	if err := updateActiveFile(activeFile, []byte(file.GetPlainName(flag.Name))); err != nil {
		log.Errorf(log.ProfSet, "Update active file err:%v", err)
		return fmt.Errorf("Update active file err:%v", err)

	}
	if len(config.KeenTune.TargetIP) == 1 {
		log.Infof(log.ProfSet, "%v Set %v successfully: %v", utils.ColorString("green", "[OK]"), flag.Name, strings.TrimPrefix(sucInfos[0], "target 1 apply result: "))
		return nil
	}

	log.Infof(log.ProfSet, "%v Set %v successfully. ", utils.ColorString("green", "[OK]"), flag.Name)
	for _, detail := range sucInfos {
		log.Infof(log.ProfSet, "\n\t%v", detail)
	}

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
	// step1. rollback the target machine
	detailInfo, allSuccess := m.Rollback(log.ProfSet)
	if !allSuccess {
		return fmt.Errorf("rollback details:\n%v", detailInfo)
	}

	// step2. clear the active file
	fileName := config.GetProfileWorkPath("active.conf")
	if err := updateActiveFile(fileName, []byte{}); err != nil {
		return fmt.Errorf("update active file failed, err:%v", err)
	}

	backupReq := utils.Parse2Map("data", configInfo)
	if backupReq == nil || len(backupReq) == 0 {
		return fmt.Errorf("backup info is null")
	}
	// step3. backup the target machine
	detailInfo, allSuccess = m.Backup(log.ProfSet, backupReq)
	if !allSuccess {
		return fmt.Errorf("backup details:\n%v", detailInfo)
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

	retRequest := map[string]interface{}{}
	retRequest["data"] = resultMap
	retRequest["resp_ip"] = respIP
	retRequest["resp_port"] = config.KeenTune.Port

	return retRequest, nil
}

func setConfiguration(request map[string]interface{}) ([]string, string, error) {
	wg := sync.WaitGroup{}
	var applyResult = make(map[int]Result)
	for index, ip := range config.KeenTune.TargetIP {
		wg.Add(1)
		go set(request, &wg, applyResult, index+1, ip)
	}

	wg.Wait()

	return analysisApplyResults(applyResult)
}

func analysisApplyResults(applyResult map[int]Result) ([]string, string, error) {
	var failedInfo string
	var successInfo []string
	// add detail info in order
	for index, _ := range config.KeenTune.TargetIP {
		id := index + 1
		if !applyResult[id].Success {
			failedInfo += applyResult[id].Info
			continue
		}
		successInfo = append(successInfo, applyResult[id].Info)
	}

	failedInfo = strings.TrimSuffix(failedInfo, ";")

	if len(successInfo) == 0 {
		return nil, failedInfo, fmt.Errorf("all failed, details:%v", successInfo)
	}

	if len(successInfo) != len(config.KeenTune.TargetIP) {
		return successInfo, failedInfo, fmt.Errorf("partial failed")
	}
	return successInfo, "", nil
}

func set(request map[string]interface{}, wg *sync.WaitGroup, applyResult map[int]Result, index int, ip string) {
	defer func() {
		wg.Done()
		config.IsInnerApplyRequests[index] = false
	}()
	uri := fmt.Sprintf("%s:%s/configure", ip, config.KeenTune.TargetPort)
	resp, err := http.RemoteCall("POST", uri, utils.ConcurrentSecurityMap(request, []string{"target_id"}, []interface{}{index}))
	if err != nil {
		applyResult[index] = Result{
			Info:    fmt.Sprintf("target %v apply remote call: %v;", index, err),
			Success: false,
		}
		return
	}

	setResult, _, err := m.GetApplyResult(resp, index)
	if err != nil {
		applyResult[index] = Result{
			Info:    fmt.Sprintf("target %v get apply result: %v;", index, err),
			Success: false,
		}
		return
	}

	applyResult[index] = Result{
		Info:    fmt.Sprintf("target %v apply result: %v", index, setResult),
		Success: true,
	}
}

func updateActiveFile(fileName string, info []byte) error {
	if err := ioutil.WriteFile(fileName, info, os.ModePerm); err != nil {
		return err
	}

	return nil
}

