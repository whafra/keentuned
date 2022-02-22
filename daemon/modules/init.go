package modules

import (
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"keentune/daemon/common/utils/http"
	"time"
)

const (
	baseOpt    = "base"
	processOpt = "process"
)

func (tuner *Tuner) init() error {
	err := tuner.initParams()
	if err != nil {
		return fmt.Errorf("init param %v", err)
	}

	err = tuner.rollback()
	if err != nil {
		return fmt.Errorf("rollback before :\n%v", tuner.rollbackDetail)
	}

	err = tuner.getConfigure()
	if err != nil {
		return fmt.Errorf("get configure:\n%v", tuner.rollbackDetail)
	}

	log.Debugf(tuner.logName, "Step%v. apply baseline configuration details: %v", tuner.Step+1, tuner.applySummary)

	err = tuner.backup()
	if err != nil {
		return fmt.Errorf("backup %v", tuner.backupDetail)
	}

	err = tuner.baseline()
	if err != nil {
		return err
	}

	err = tuner.brainInit()
	if err != nil {
		return err
	}

	if tuner.isInterrupted() {
		log.Infof(tuner.logName, "Tuning interrupted after step%v, [init] finish.", tuner.Step)
		return fmt.Errorf("tuning is interrupted")
	}

	return nil
}

func (tuner *Tuner) dump(option string) {
	var suffix string
	switch option {
	case baseOpt:
		if !config.KeenTune.DumpConf.BaseDump || tuner.isSensitize {
			return
		}
		suffix = "_base.json"
	case processOpt:
		if !config.KeenTune.DumpConf.ExecDump {
			return
		}
		suffix = fmt.Sprintf("_round_%v.json", tuner.Iteration)
	default:
		return
	}

	for index := range tuner.Group {
		if config.KeenTune.DumpConf.BaseDump {
			tuner.Group[index].Dump.Round = -1
		}

		tuner.Group[index].Dump.Score = tuner.benchScore
		err := tuner.Group[index].Dump.Save(tuner.Name, fmt.Sprintf("_group%v%v", index+1, suffix))
		if err != nil {
			log.Warnf(tuner.logName, "dump %v failed, %v", option, err)
		}
	}
}
func (tuner *Tuner) dumpBest() error {
	if !config.KeenTune.DumpConf.BestDump {
		return nil
	}

	err := tuner.getBest()
	if err != nil {
		return err
	}

	err = tuner.parseBestParam()
	if err != nil {
		return err
	}

	suffix := "_best.json"

	for index := range tuner.Group {
		err = tuner.Group[index].Dump.Save(tuner.Name, fmt.Sprintf("_group%v%v", index+1, suffix))
		if err != nil {
			log.Warnf(tuner.logName, "dump best.json failed, %v", err)
		}
	}

	log.Infof(tuner.logName, "Step%v. Best configuration dump to [%v/parameter/%v/%v] successfully.\n", tuner.IncreaseStep(), config.KeenTune.DumpConf.DumpHome, tuner.Name, tuner.Name+"group**_best.json")
	return nil
}

func (tuner *Tuner) baseline() error {
	success, _, err := tuner.Benchmark.SendScript(&tuner.timeSpend.send)
	if err != nil || !success {
		return fmt.Errorf("send script file  result: %v, details:%v", success, err)
	}

	_, tuner.benchScore, tuner.benchSummary, err = tuner.RunBenchmark(config.KeenTune.BaseRound)
	if err != nil {
		if err.Error() == "get benchmark is interrupted" {
			log.Infof(tuner.logName, "Tuning interrupted after step%v, [baseline benchmark] stopped.", tuner.Step)
			return fmt.Errorf("run benchmark interrupted")
		}
		return fmt.Errorf("tuning execute baseline benchmark:%v", err)
	}

	if !tuner.isSensitize {
		log.Infof(tuner.logName, "Step%v. Run benchmark as baseline:%v", tuner.IncreaseStep(), tuner.benchSummary)
		tuner.dump(baseOpt)
	}

	return nil
}

func (tuner *Tuner) brainInit() error {
	err := tuner.getBrainInitParams()
	if err != nil {
		return fmt.Errorf("get brain param: %v", err)
	}

	var requireConf = make(map[string]interface{})

	requireConf["algorithm"] = tuner.Algorithm
	requireConf["iteration"] = tuner.MAXIteration
	requireConf["name"] = tuner.Name
	requireConf["type"] = tuner.Flag
	requireConf["parameters"] = tuner.BrainParam
	requireConf["baseline_score"] = tuner.benchScore

	start := time.Now()

	url := config.KeenTune.BrainIP + ":" + config.KeenTune.BrainPort + "/init"
	err = http.ResponseSuccess("POST", url, requireConf)
	if err != nil {
		return fmt.Errorf("remote call [init] failed: %v", err)
	}

	timeCost := utils.Runtime(start)
	tuner.timeSpend.init += timeCost.Count

	if !tuner.isSensitize {
		log.Infof(tuner.logName, "Step%v. AI Engine is ready.\n", tuner.IncreaseStep())
	}
	return nil
}

func (tuner *Tuner) rollback() error {
	var sucCount int
	var retResult string
	for i, target := range tuner.Group {
		go func(i int, target Group) {
			result, allSuc := target.Rollback()
			retResult += result
			if allSuc {
				sucCount++
			} else {
				log.Warnf(tuner.logName, "rollback failed, %v", result)
			}
		}(i, target)
	}

	if sucCount != len(tuner.Group) {
		return fmt.Errorf("failure occur")
	}

	tuner.rollbackDetail = retResult
	return nil
}

func (tuner *Tuner) backup() error {
	var sucCount int
	var retResult string
	for i, target := range tuner.Group {
		go func(i int, target Group) {
			result, allSuc := target.Backup(target.MergedParam)
			retResult += result
			if allSuc {
				sucCount++
			}
		}(i, target)
	}

	if sucCount != len(tuner.Group) {
		return fmt.Errorf("failure occur")
	}

	tuner.backupDetail = retResult

	return nil
}

