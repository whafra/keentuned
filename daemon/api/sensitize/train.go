package sensitize

import (
	com "keentune/daemon/api/common"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"keentune/daemon/common/utils/http"
	m "keentune/daemon/modules"
	"encoding/json"
	"fmt"
	"os"
	"io/ioutil"
)

type TrainFlag struct {
	Output string
	Data   string
	Trials int
	Force  bool
	Log    string
}

// Train run sensitize train service
func (s *Service) Train(flags TrainFlag, reply *string) error {
	if com.SystemRun {
		log.Errorf("", "An instance is running. You can wait the process finish or run \"keentune %v stop\" for trying a new job again if you want give up the old job.", com.GetRunningTask())
		return fmt.Errorf("Train failed, an instance is running. You can wait the process finish or run \"keentune %v stop\" and try a new job again, if you want give up the old job.", com.GetRunningTask())
	}

	go runTrain(flags)
	return nil
}

func runTrain(flags TrainFlag) {
	log.SensitizeTrain = "sensitize train" + ":" + flags.Log
	com.SystemRun = true
	com.IsSensitizing = true
	ioutil.WriteFile(flags.Log, []byte{}, os.ModePerm)
	defer func() {
		log.ClearCliLog(log.SensitizeTrain)
		config.ProgramNeedExit <- true
		<-config.ServeFinish
		com.SystemRun = false
		com.IsSensitizing = false
	}()

	log.Infof(log.SensitizeTrain, "Step1. Sensitize train data [%v] start.", flags.Data)
	go com.HeartbeatCheck()

	if !isTrainFlagsRightful(flags) {
		log.Errorf(log.SensitizeTrain, "check train options failed")
		return
	}

	if err := initiateSensitization(&flags); err != nil {
		return
	}

	log.Infof(log.SensitizeTrain, "\nStep2. Initiate sensitization success.\n")

	resultString, resultMap, err := getSensitivityResult()
	if err != nil {
		log.Errorf(log.SensitizeTrain, "Get sensitivity result failed, err:%v", err)
		return
	}

	log.Infof(log.SensitizeTrain, "Step3. Get sensitive parameter identification results successfully, and the details are as follows.%v", resultString)
	

	if err = dumpSensitivityResult(resultMap, flags.Output); err != nil {
		return
	}

	log.Infof(log.SensitizeTrain, "\nStep4. Dump sensitivity result to %v successfully, and \"sensitize train\" finish.\n", fmt.Sprintf("%s/sensi-%s.json", m.GetSensitizePath(), flags.Output))	
}

func initiateSensitization(flags *TrainFlag) error {
	uri := config.KeenTune.BrainIP + ":" + config.KeenTune.BrainPort + "/sensitize"

	ip, err := utils.GetExternalIP()
	if err != nil {
		log.Errorf(log.SensitizeTrain, "get local external ip err:%v", err)
		return err
	}

	reqInfo := map[string]interface{}{
		"data":      flags.Data,
		"resp_ip":   ip,
		"resp_port": config.KeenTune.Port,
		"trials":    flags.Trials,
	}

	err = http.ResponseSuccess("POST", uri, reqInfo)
	if err != nil {
		log.Errorf(log.SensitizeTrain, "RemoteCall [POST] sensitize err:%v", err)
		return err
	}

	return nil
}

func getSensitivityResult() (string, map[string]interface{}, error) {
	var sensitizeParams struct {
		Success bool          `json:"suc"`
		Result  []m.Parameter `json:"result"`
		Msg     interface{}   `json:"msg"`
	}

	config.IsInnerRequests = true

	select {
	case resultBytes := <-config.SensitizeReusltChan :
		log.Debugf(log.SensitizeTrain, "get sensitivity result:%s", resultBytes)
		if len(resultBytes) == 0 {
			return "", nil, fmt.Errorf("get sensitivity result is nil")
		}

		if err := json.Unmarshal(resultBytes, &sensitizeParams); err != nil {
			return "", nil, err
		}
	}

	if !sensitizeParams.Success {
		return "", nil, fmt.Errorf("error msg:%v", sensitizeParams.Msg)
	}

	domainMap := make(map[string][]map[string]interface{})
	resultMap := make(map[string]interface{})
	var resultString string
	if len(sensitizeParams.Result) >0 {
		resultString += fmt.Sprintf("\n\t| %v | %v |", "parameter name", "sensitivity ratio")
	}

	for _, param := range sensitizeParams.Result {
		paramInfo := map[string]interface{}{
			param.ParaName: map[string]interface{}{"weight": param.Weight},
		}

		resultString += fmt.Sprintf("\n\t| %v \t|  %v \t|", param.ParaName, param.Weight)
		domainMap[param.DomainName] = append(domainMap[param.DomainName], paramInfo)
	}

	for domain, paramSlice := range domainMap {
		paramMap := make(map[string]interface{})
		for _, info := range paramSlice {
			for name, value := range info {
				paramMap[name] = value
			}
		}
		resultMap[domain] = paramMap
	}

	return resultString, resultMap, nil
}

func isTrainFlagsRightful(flag TrainFlag) bool {
	if !com.IsDataNameUsed(flag.Data) {
		log.Errorf(log.SensitizeTrain, "check data name [%v] not exists", flag.Data)
		return false
	}

	fileName := fmt.Sprintf("%s/sensi-%s.json", m.GetSensitizePath(), flag.Output)
	if file.IsPathExist(fileName)&&!flag.Force {
		log.Errorf(log.SensitizeTrain, "output file [%v] exist and you have given up to overwrite it", flag.Output)
		return false
	}
	return true
}

func dumpSensitivityResult(resultMap map[string]interface{}, recordName string) error {
	fileName := "sensi-" + recordName + ".json"
	if err := file.Dump2File(m.GetSensitizePath(), fileName, resultMap); err != nil {
		log.Errorf(log.SensitizeTrain, "dump sensitivity result to file [%v] err:[%v] ", fileName, err)
		return err
	}

	return nil
}
