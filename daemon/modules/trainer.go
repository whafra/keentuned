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
	"os"
	"strings"
)

// Tuner define a tuning job include Algorithm, Benchmark, Group
type Trainer struct {
	Data       string
	Job        string
	Trials     int
	Config     string
	BenchRound int
}

// Tune : training main process
func (tuner *Tuner) Train() {
	var err error
	tuner.logName = log.SensitizeTrain
	if err = tuner.CreateTrainJob(); err != nil {
		log.Errorf(log.SensitizeTrain, "create sensitize train job failed: %v", err)
		return
	}

	defer func() { tuner.parseTuningError(err) }()

	if err = tuner.initiateSensitization(); err != nil {
		return
	}

	tuner.copyKnobsFile()

	log.Infof(log.SensitizeTrain, "\nStep2. Initiate sensitization success.\n")

	resultString, resultMap, err := tuner.getSensitivityResult()
	if err != nil {
		log.Errorf(log.SensitizeTrain, "Get sensitivity result failed, err:%v", err)
		return
	}

	log.Infof(log.SensitizeTrain, "Step3. Get sensitive parameter identification results successfully, and the details are as follows.%v", resultString)

	if err = tuner.dumpSensitivityResult(resultMap, tuner.Job); err != nil {
		return
	}

	log.Infof(log.SensitizeTrain, "\nStep4. Dump sensitivity result to %v successfully, and \"sensitize train\" finish.\n", fmt.Sprintf("%s/sensi-%s.json", config.GetSensitizePath(""), tuner.Job))
}

func (tuner *Tuner) CreateTrainJob() error {
	log := fmt.Sprintf("%v/%v-%v.log", "/var/log/keentune", "keentuned-sensitize-train", tuner.Job)

	jobInfo := []string{
		tuner.Job, tuner.StartTime.Format(Format), NA, NA, fmt.Sprint(tuner.Trials), Run,
		fmt.Sprint(tuner.BenchRound), log, config.GetSensitizeWorkPath(tuner.Job), tuner.Algorithm, tuner.Data,
	}
	return file.Insert(getSensitizeJobFile(), jobInfo)
}

func (tuner *Tuner) initiateSensitization() error {
	uri := config.KeenTune.BrainIP + ":" + config.KeenTune.BrainPort + "/sensitize"

	ip, err := utils.GetExternalIP()
	if err != nil {
		log.Errorf(log.SensitizeTrain, "get local external ip err:%v", err)
		return err
	}

	reqInfo := map[string]interface{}{
		"data":      tuner.Data,
		"resp_ip":   ip,
		"resp_port": config.KeenTune.Port,
		"trials":    tuner.Trials,
	}

	err = http.ResponseSuccess("POST", uri, reqInfo)
	if err != nil {
		log.Errorf(log.SensitizeTrain, "RemoteCall [POST] sensitize err:%v", err)
		return err
	}

	return nil
}

func (tuner *Tuner) getSensitivityResult() (string, map[string]interface{}, error) {

	var sensitizeParams struct {
		Success bool        `json:"suc"`
		Head    string      `json:"head"`
		Result  [][]float64 `json:"result"`
		Msg     interface{} `json:"msg"`
	}

	config.IsInnerSensitizeRequests[1] = true
	defer func() { config.IsInnerSensitizeRequests[1] = false }()
	select {
	case resultBytes := <-config.SensitizeResultChan:
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

	sensiResultCsv := fmt.Sprintf("%v/sensi_result.csv", config.GetSensitizePath(tuner.Job))
	sensiResultHeader := strings.Split(sensitizeParams.Head, ",")
	if len(sensitizeParams.Result) == 0 || len(sensitizeParams.Result[0]) != len(sensiResultHeader) {
		return "", nil, fmt.Errorf("error msg:Header does not match param")
	}

	if !file.IsPathExist(sensiResultCsv) {
		err := file.CreatCSV(sensiResultCsv, sensiResultHeader)
		if err != nil {
			fmt.Printf("%v create sensitize jobs csv file: %v", utils.ColorString("red", "[ERROR]"), err)
			os.Exit(1)
		}
	}
	var resultSlice [][]string
	resultSlice = append(resultSlice, sensiResultHeader)

	for _, paramSlice := range sensitizeParams.Result {
		var endInfo []string
		for _, param := range paramSlice {
			endInfo = append(endInfo, fmt.Sprint(param))
		}
		resultSlice = append(resultSlice, endInfo)
		file.Insert(sensiResultCsv, endInfo)
	}
	return utils.FormatInTable(resultSlice), nil, nil

}

func (tuner *Tuner) dumpSensitivityResult(resultMap map[string]interface{}, recordName string) error {
	fileName := "sensi-" + recordName + ".json"
	if err := file.Dump2File(config.GetSensitizePath(""), fileName, resultMap); err != nil {
		log.Errorf(log.SensitizeTrain, "dump sensitivity result to file [%v] err:[%v] ", fileName, err)
		return err
	}

	return nil
}

func (tuner *Tuner) copyKnobsFile() error {
	tuneknobsFile := fmt.Sprintf("%v/knobs.json", config.GetTuningPath(tuner.Data))
	sensiknobsFile := fmt.Sprintf("%v/knobs.json", config.GetSensitizePath(tuner.Job))

	input, err := ioutil.ReadFile(tuneknobsFile)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = ioutil.WriteFile(sensiknobsFile, input, 0666)
	if err != nil {
		fmt.Println("Error creating", sensiknobsFile)
		fmt.Println(err)
		return err
	}
	return nil
}
