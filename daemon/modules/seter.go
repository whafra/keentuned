package modules

import (
	"fmt"
	"io/ioutil"
	com "keentune/daemon/api/common"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"keentune/daemon/common/utils/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Tuner define a tuning job include Algorithm, Benchmark, Group
// type Seter struct {
// 	Name     string
// 	Group    []bool
// 	ConfFile []string
// }

type Seter struct {
	Name     string
	Group    []bool
	ConfFile []string
}

type ResultProfileSet struct {
	Info    string
	Success bool
}

// Tune : tuning main process
func (tuner *Tuner) Set() {
	var err error
	tuner.logName = log.ProfSet
	defer func() {
		if err != nil {
			tuner.endSet()
			tuner.parseSetingError(err)
		}
	}()

	if err = tuner.initProfiles(); err != nil {
		err = fmt.Errorf("[%v] prepare for profile set: %v", utils.ColorString("red", "ERROR"), err)
		return
	}

	configFileALL, err := tuner.checkProfilePath()
	if err != nil {
		log.Errorf(log.ProfSet, "Check file err:%v", err)
		return
	}
	fmt.Println(configFileALL)

	requestInfoAll, err := tuner.getConfigParamInfo(configFileALL)
	if err != nil {
		log.Errorf(log.ProfSet, "Get request info from specified file [%v] err:%v", tuner.Seter.ConfFile[0], err)
		return
	}
	fmt.Println(requestInfoAll)

	if err := tuner.prepareBeforeSet(requestInfoAll); err != nil {
		log.Errorf(log.ProfSet, "Prepare for Set err:%v", err)
		return
	}

	sucInfos, failedInfo, err := tuner.setConfiguration(requestInfoAll)
	if err != nil {
		log.Errorf(log.ProfSet, "Set failed:%v, details:%v", err, failedInfo)
		return
	}

	activeFile := config.GetProfileWorkPath("active.conf")
	if err := appendActiveFile(activeFile, []byte(file.GetPlainName("123"))); err != nil {
		log.Errorf(log.ProfSet, "Update active file err:%v", err)
		return
	}

	for groupIndex, v := range tuner.Seter.Group {
		if v {
			log.Infof(log.ProfSet, "%v Set %v successfully: %v", utils.ColorString("green", "[OK]"), tuner.Seter.ConfFile[groupIndex], strings.TrimPrefix(sucInfos[groupIndex], "target 1 apply result: "))
			log.Infof(log.ProfSet, "%v Set %v successfully. ", utils.ColorString("green", "[OK]"), tuner.Seter.ConfFile[groupIndex])
		}
	}

	for _, detail := range sucInfos {
		log.Infof(log.ProfSet, "\n\t%v", detail)
	}
	return
}

func (tuner *Tuner) checkProfilePath() (map[int]string, error) {

	filePathAll := make(map[int]string) //key为groupNo，value为.conf
	for groupIndex, v := range tuner.Seter.Group {
		if v {
			filePath := com.GetProfilePath(tuner.Seter.ConfFile[groupIndex])
			if filePath != "" {
				filePathAll[groupIndex] = filePath
			} else {
				return nil, fmt.Errorf("find the configuration file [%v] neither in[%v] nor in [%v]", tuner.Seter.ConfFile[groupIndex], fmt.Sprintf("%s/profile", config.KeenTune.Home), fmt.Sprintf("%s/profile", config.KeenTune.DumpHome))
			}
		}
	}
	return filePathAll, nil

}

func (tuner *Tuner) prepareBeforeSet(configInfoAll map[int]map[string]interface{}) error {
	// step1. rollback the target machine
	err := tuner.rollback_profile_set()
	if err != nil {
		return fmt.Errorf("rollback details:\n%v", err)
	}

	// step2. clear the active file
	fileName := config.GetProfileWorkPath("active.conf")
	if err := UpdateActiveFile(fileName, []byte{}); err != nil {
		return fmt.Errorf("update active file failed, err:%v", err)
	}

	for _, configInfo := range configInfoAll {
		backupReq := utils.Parse2Map("data", configInfo)
		//if backupReq == nil || len(backupReq) == 0 {
		if len(backupReq) == 0 {
			return fmt.Errorf("backup info is null")
		}
	}
	// step3. backup the target machine
	err = tuner.backup_profile_set()
	if err != nil {
		return fmt.Errorf("backup details:\n%v", err)
	}
	return nil
}

func (tuner *Tuner) getConfigParamInfo(configFileALL map[int]string) (map[int]map[string]interface{}, error) {

	retRequestAll := map[int]map[string]interface{}{}
	retRequest := map[string]interface{}{}
	for groupIndex, configFile := range configFileALL {

		resultMap, err := file.ConvertConfFileToJson(configFile)
		if err != nil {
			return nil, fmt.Errorf("convert file :%v err:%v", configFile, err)
		}

		respIP, err := utils.GetExternalIP()
		if err != nil {
			return nil, fmt.Errorf("run benchmark get real keentuned ip err: %v", err)
		}

		retRequest["data"] = resultMap
		retRequest["resp_ip"] = respIP
		retRequest["resp_port"] = config.KeenTune.Port
		retRequestAll[groupIndex] = retRequest

	}
	return retRequestAll, nil
}

func (tuner *Tuner) setConfiguration(requestAll map[int]map[string]interface{}) ([]string, string, error) {
	wg := sync.WaitGroup{}
	var applyResult = make(map[int]ResultProfileSet)

	//groupIndex为target-group-x   x= groupIndex + 1
	for groupIndex, request := range requestAll {
		for _, target := range tuner.Group {
			if target.GroupNo == groupIndex+1 {
				for index, ip := range target.IPs {
					wg.Add(1)
					go tuner.set(request, &wg, applyResult, index+1, ip)
				}
			}
		}
	}

	wg.Wait()

	return tuner.analysisApplyResults(applyResult)
}

func (tuner *Tuner) analysisApplyResults(applyResult map[int]ResultProfileSet) ([]string, string, error) {
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

func (tuner *Tuner) set(request map[string]interface{}, wg *sync.WaitGroup, applyResult map[int]ResultProfileSet, index int, ip string) {
	defer func() {
		wg.Done()
		config.IsInnerApplyRequests[index] = false
	}()
	uri := fmt.Sprintf("%s:%s/configure", ip, config.KeenTune.TargetPort)
	resp, err := http.RemoteCall("POST", uri, utils.ConcurrentSecurityMap(request, []string{"target_id", "readonly"}, []interface{}{index, false}))
	if err != nil {
		applyResult[index] = ResultProfileSet{
			Info:    fmt.Sprintf("target %v apply remote call: %v;", index, err),
			Success: false,
		}
		return
	}

	setResult, _, err := GetApplyResult(resp, index)
	if err != nil {
		applyResult[index] = ResultProfileSet{
			Info:    fmt.Sprintf("target %v get apply result: %v;", index, err),
			Success: false,
		}
		return
	}

	applyResult[index] = ResultProfileSet{
		Info:    fmt.Sprintf("target %v apply result: %v", index, setResult),
		Success: true,
	}
}

// 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func appendActiveFile(fileName string, info []byte) error {
	// if err := ioutil.WriteFile(fileName, info, os.ModePerm); err != nil {
	// 	return err
	// }
	var file *os.File
	var err error
	if Exists(fileName) {
		//使用追加模式打开文件
		file, err = os.OpenFile(fileName, os.O_APPEND, 0666)
		if err != nil {
			fmt.Println("Open file err =", err)
			return err
		}
	} else {
		file, err = os.Create(fileName) //创建文件
		if err != nil {
			fmt.Println("file create fail")
			return err
		}
	}
	defer file.Close()

	n, err := file.WriteString(string(info))
	if err != nil {
		fmt.Println("Write file err =", err)
		return err
	}
	fmt.Println("Write file success, n =", n)
	return nil
}

func UpdateActiveFile(fileName string, info []byte) error {
	if err := ioutil.WriteFile(fileName, info, os.ModePerm); err != nil {
		return err
	}

	return nil
}

func (tuner *Tuner) parseSetingError(err error) {
	if err == nil {
		return
	}

	tuner.rollback()

	if strings.Contains(err.Error(), "interrupted") {
		log.Infof(tuner.logName, "profile optimization job abort!")
		return
	}
	log.Infof(tuner.logName, "%v", err)
}

func (tuner *Tuner) endSet() {
	start := time.Now()
	timeCost := utils.Runtime(start)
	tuner.timeSpend.end += timeCost.Count

	totalTime := utils.Runtime(tuner.StartTime).Count.Seconds()

	if totalTime == 0.0 || !tuner.Verbose {
		return
	}

	tuner.setTimeSpentDetail(totalTime)
}
