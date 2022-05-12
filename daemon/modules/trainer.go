package modules

import (
	"encoding/json"
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"keentune/daemon/common/utils/http"
	"strings"
	"time"
)

// Tuner define a tuning job include Algorithm, Benchmark, Group
type Trainer struct {
	Name          string
	Algorithm     string // 使用的算法
	MAXIteration  int    // 最大执行轮次
	Iteration     int    // 当前轮次
	StartTime     time.Time
	Benchmark     Benchmark
	timeSpend     TimeSpend
	ParamConf     config.DBLMap
	Verbose       bool
	Step          int    // tuning process steps
	isSensitize   bool   // sensitive parameter identification mark
	Flag          string // command flag, enum: "collect", "tuning"
	logName       string
	Data          string
	Job           string
	Trials        int
	Config        string
	feedbackScore map[string][]float32
	benchScore    map[string]ItemDetail
	Group         []Group
	BrainParam    []Parameter
	ReadConfigure bool
	implyDetail
	bestInfo    Configuration
	allowUpdate bool
}

// Tune : tuning main process
func (trainer *Trainer) Train() {
	var err error
	trainer.logName = log.SensitizeTrain
	if err = trainer.CreateTrainJob(); err != nil {
		log.Errorf(log.SensitizeTrain, "create sensitize train job failed: %v", err)
		return
	}

	defer func() {
		if err != nil {
			trainer.end()
			trainer.parseTuningError(err)
		}
	}()

	if err = trainer.initiateSensitization(); err != nil {
		return
	}

	log.Infof(log.SensitizeTrain, "\nStep2. Initiate sensitization success.\n")

	resultString, resultMap, err := trainer.getSensitivityResult()
	if err != nil {
		log.Errorf(log.SensitizeTrain, "Get sensitivity result failed, err:%v", err)
		return
	}

	log.Infof(log.SensitizeTrain, "Step3. Get sensitive parameter identification results successfully, and the details are as follows.%v", resultString)

	if err = trainer.dumpSensitivityResult(resultMap, trainer.Job); err != nil {
		return
	}

	log.Infof(log.SensitizeTrain, "\nStep4. Dump sensitivity result to %v successfully, and \"sensitize train\" finish.\n", fmt.Sprintf("%s/sensi-%s.json", config.GetSensitizePath(trainer.Job), trainer.Job))
}

func (trainer *Trainer) CreateTrainJob() error {
	//cmd := fmt.Sprintf("keentune sensitize train --data %v --job %v --trials %v --config %v", trainer.Data, trainer.Job, trainer.Trials, trainer.Config)

	log := fmt.Sprintf("%v/%v.log", "/var/log/keentune", trainer.Job)

	jobInfo := []string{
		trainer.Job, NA, NA, NA, fmt.Sprint(trainer.Trials), Run,
		"0", log, config.GetSensitizeWorkPath(trainer.Job), trainer.Algorithm, trainer.Data,
	}
	return file.Insert(getSensitizeJobFile(), jobInfo)
}

func (trainer *Trainer) initiateSensitization() error {
	uri := config.KeenTune.BrainIP + ":" + config.KeenTune.BrainPort + "/sensitize"

	ip, err := utils.GetExternalIP()
	if err != nil {
		log.Errorf(log.SensitizeTrain, "get local external ip err:%v", err)
		return err
	}

	reqInfo := map[string]interface{}{
		"data":      trainer.Data,
		"resp_ip":   ip,
		"resp_port": config.KeenTune.Port,
		"trials":    trainer.Trials,
	}

	err = http.ResponseSuccess("POST", uri, reqInfo)
	if err != nil {
		log.Errorf(log.SensitizeTrain, "RemoteCall [POST] sensitize err:%v", err)
		return err
	}

	return nil
}

func (trainer *Trainer) getSensitivityResult() (string, map[string]interface{}, error) {
	var sensitizeParams struct {
		Success bool        `json:"suc"`
		Result  []Parameter `json:"result"`
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

	domainMap := make(map[string][]map[string]interface{})
	resultMap := make(map[string]interface{})
	var resultSlice [][]string
	if len(sensitizeParams.Result) > 0 {
		resultSlice = append(resultSlice, []string{"parameter name", "sensitivity ratio"})
	}

	for _, param := range sensitizeParams.Result {
		paramInfo := map[string]interface{}{
			param.ParaName: map[string]interface{}{"weight": param.Weight},
		}

		resultSlice = append(resultSlice, []string{param.ParaName, fmt.Sprint(param.Weight)})
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

	return utils.FormatInTable(resultSlice), resultMap, nil
}

func (trainer *Trainer) dumpSensitivityResult(resultMap map[string]interface{}, recordName string) error {
	fileName := "sensi-" + recordName + ".json"
	if err := file.Dump2File(config.GetSensitizePath(fileName), fileName, resultMap); err != nil {
		log.Errorf(log.SensitizeTrain, "dump sensitivity result to file [%v] err:[%v] ", fileName, err)
		return err
	}

	return nil
}

func (trainer *Trainer) parseTuningError(err error) {

	if err == nil {
		trainer.updateStatus(Finish)
		return
	}

	//trainer.rollback()
	if strings.Contains(err.Error(), "interrupted") {
		trainer.updateStatus(Stop)
		log.Infof(trainer.logName, "parameter optimization job abort!")
		return
	}

	trainer.updateStatus(Err)
	log.Infof(trainer.logName, "%v", err)

}

func (trainer *Trainer) end() {

	start := time.Now()
	http.RemoteCall("GET", config.KeenTune.BrainIP+":"+config.KeenTune.BrainPort+"/end", nil)
	timeCost := utils.Runtime(start)
	trainer.timeSpend.end += timeCost.Count

	totalTime := utils.Runtime(trainer.StartTime).Count.Seconds()

	if totalTime == 0.0 || !trainer.Verbose {
		return
	}

	trainer.setTimeSpentDetail(totalTime)

	var endInfo = make(map[int]interface{})

	if trainer.Flag == "tuning" {
		endInfo[tuneEndIdx] = start.Format(Format)
		endInfo[tuneCostIdx] = endTime(int64(totalTime))
	}

	trainer.updateJob(endInfo)
}

func (trainer *Trainer) setTimeSpentDetail(totalTime float64) {
	initRatio := fmt.Sprintf("%.2f%%", trainer.timeSpend.init.Seconds()*100/totalTime)
	applyRatio := fmt.Sprintf("%.2f%%", trainer.timeSpend.apply.Seconds()*100/totalTime)
	acquireRatio := fmt.Sprintf("%.2f%%", trainer.timeSpend.acquire.Seconds()*100/totalTime)
	benchmarkRatio := fmt.Sprintf("%.2f%%", trainer.timeSpend.benchmark.Seconds()*100/totalTime)
	feedbackRatio := fmt.Sprintf("%.2f%%", trainer.timeSpend.feedback.Seconds()*100/totalTime)

	var detailSlice [][]string
	header := []string{"Process", "Execution Count", "Total Time", "The Share of Total Time"}
	detailSlice = append(detailSlice, header)

	initTime := fmt.Sprintf("%.3fs", trainer.timeSpend.init.Seconds())
	detailSlice = append(detailSlice, []string{"init", "1", initTime, initRatio})

	applyRound := fmt.Sprint(trainer.MAXIteration + 2)
	applyTime := fmt.Sprintf("%.3fs", trainer.timeSpend.apply.Seconds())
	detailSlice = append(detailSlice, []string{"apply", applyRound, applyTime, applyRatio})

	maxRound := fmt.Sprint(trainer.MAXIteration)
	acquireTime := fmt.Sprintf("%.3fs", trainer.timeSpend.acquire.Seconds())
	detailSlice = append(detailSlice, []string{"acquire", maxRound, acquireTime, acquireRatio})

	benchRound := fmt.Sprint(trainer.MAXIteration*config.KeenTune.ExecRound + config.KeenTune.BaseRound + config.KeenTune.AfterRound)
	benchTime := fmt.Sprintf("%.3fs", trainer.timeSpend.benchmark.Seconds())
	detailSlice = append(detailSlice, []string{"benchmark", benchRound, benchTime, benchmarkRatio})

	feedbackTime := fmt.Sprintf("%.3fs", trainer.timeSpend.feedback.Seconds())
	detailSlice = append(detailSlice, []string{"feedback", maxRound, feedbackTime, feedbackRatio})

	trainer.timeSpend.detailInfo = utils.FormatInTable(detailSlice)
}
