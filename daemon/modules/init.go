package modules

import (
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"keentune/daemon/common/utils/http"
	"time"
)

func (tuner *Tuner) init() error {

	result, allSuccess := tuner.rollback()
	if !allSuccess {
		return fmt.Errorf("Rollback before %v:\n%v", tuner.Flag, result)
	}

	emptyConf := AssembleParams(tuner.ParamConf)
	if emptyConf == nil {
		return fmt.Errorf("read or assemble parameter failed")
	}

	var implyApplyResults string
	var err error
	implyApplyResults, tuner.BaseConfiguration, err = emptyConf.Apply(&tuner.timeSpend.apply, true)
	if err != nil {
		return fmt.Errorf("baseline apply configuration failed: %v, details: %v", err, implyApplyResults)
	}

	if err = backup(tuner.logName, emptyConf); err != nil {
		return err
	}

	log.Debugf(tuner.logName, "Step%v. apply baseline configuration details: %v", tuner.Step+1, implyApplyResults)

	success, _, err := tuner.Benchmark.SendScript(&tuner.timeSpend.send)
	if err != nil || !success {
		return fmt.Errorf("send script file  result: %v, details:%v", success, err)
	}

	_, scoreResult, implyBenchResult, err := tuner.Benchmark.RunBenchmark(config.KeenTune.BaseRound, &tuner.timeSpend.benchmark, tuner.Verbose)
	if err != nil {
		if err.Error() == "get benchmark is interrupted" {
			log.Infof(log.ParamTune, "Tuning interrupted after step%v, [baseline benchmark] stopped.", tuner.Step)
			return fmt.Errorf("run benchmark interrupted")
		}
		return fmt.Errorf("tuning execute baseline benchmark:%v", err)
	}

	tuner.BaseConfiguration[0].Score = scoreResult
	tuner.BaseConfiguration[0].UpdateBase(emptyConf)
	if tuner.logName == log.ParamTune {
		log.Infof(tuner.logName, "Step%v. AI Engine is ready.\n", tuner.IncreaseStep())

		log.Infof(tuner.logName, "Step%v. Run benchmark as baseline:%v", tuner.IncreaseStep(), implyBenchResult)

		if config.KeenTune.DumpConf.BaseDump {
			for index := range tuner.BaseConfiguration {
				targetID := index + 1
				tuner.BaseConfiguration[index].Round = 0
				tuner.BaseConfiguration[index].Score = scoreResult
				tuner.BaseConfiguration[index].Dump(tuner.Name, fmt.Sprintf("target_%v_base.json", targetID))
			}
		}
	}

	var requireConf = make(map[string]interface{})

	requireConf["algorithm"] = tuner.Algorithm
	requireConf["iteration"] = tuner.MAXIteration
	requireConf["name"] = tuner.Name
	requireConf["type"] = tuner.Flag

	requireConf["parameters"] = emptyConf.Parameters
	requireConf["baseline_score"] = scoreResult

	start := time.Now()

	err = http.ResponseSuccess("POST", config.KeenTune.BrainIP+":"+config.KeenTune.BrainPort+"/init", requireConf)
	if err != nil {
		return fmt.Errorf("remote call [init] failed: %v", err)
	}

	timeCost := utils.Runtime(start)
	tuner.timeSpend.init += timeCost.Count

	if isInterrupted(tuner.logName) {
		log.Infof(tuner.logName, "Tuning interrupted after step%v, [init] finish.", tuner.Step)
		return fmt.Errorf("tuning is interrupted")
	}

	return nil
}

func (tuner *Tuner) rollback() (string, bool) {
	var sucCount int
	var retResult string
	for i, target := range tuner.Group {
		go func(i int, target Target) {
			result, allSuc := target.Rollback()
			retResult += result
			if allSuc {
				sucCount++
			}
		}(i, target)
	}

	if sucCount != len(tuner.Group) {
		return retResult, false
	}
	return retResult, true
}

func (tuner *Tuner) backup() (string, bool) {
	var sucCount int
	var retResult string
	for i, target := range tuner.Group {
		go func(i int, target Target) {
			result, allSuc := target.Backup()
			retResult += result
			if allSuc {
				sucCount++
			}
		}(i, target)
	}

	if sucCount != len(tuner.Group) {
		return retResult, false
	}
	return retResult, true
}
