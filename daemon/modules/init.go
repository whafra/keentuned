package modules

import (
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"keentune/daemon/common/utils/http"
	"strings"
	"sync"
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
		return fmt.Errorf("rollback:\n%v", tuner.rollbackDetail)
	}

	err = tuner.getConfigure()
	if err != nil {
		return fmt.Errorf("get configure:\n%v", err)
	}

	log.Debugf(tuner.logName, "Step%v. apply baseline configuration details: %v", tuner.Step+1, tuner.applySummary)

	err = tuner.backup()
	if err != nil {
		return fmt.Errorf("backup %v", tuner.backupFailure)
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

func (tuner *Tuner) baseline() error {
	if !tuner.isSensitize {
		log.Infof(tuner.logName, "Step%v. Run benchmark as baseline:", tuner.IncreaseStep())
	}

	for _, benchgroup := range config.KeenTune.BenchGroup {
		for _, benchip := range benchgroup.SrcIPs {
			Host := fmt.Sprintf("%s:%s", benchip, benchgroup.SrcPort)
			success, _, err := tuner.Benchmark.SendScript(&tuner.timeSpend.send, Host)
			if err != nil || !success {
				return fmt.Errorf("send script file  result: %v, details:%v", success, err)
			}
		}
	}

	var err error
	_, tuner.benchScore, tuner.benchSummary, err = tuner.RunBenchmark(config.KeenTune.BaseRound)
	if err != nil {
		if err.Error() == "get benchmark is interrupted" {
			log.Infof(tuner.logName, "Tuning interrupted after step%v, [baseline benchmark] stopped.", tuner.Step)
			return fmt.Errorf("run benchmark interrupted")
		}
		return fmt.Errorf("tuning execute baseline benchmark:%v", err)
	}

	if !tuner.isSensitize {
		log.Infof(tuner.logName, "%v", tuner.benchSummary)
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
		log.Infof(tuner.logName, "\nStep%v. AI Engine is ready.", tuner.IncreaseStep())
	}
	return nil
}

func (tuner *Tuner) rollback() error {
	return tuner.concurrent("rollback", false)
}

func (tuner *Tuner) backup() error {
	return tuner.concurrent("backup", true)
}

func (tuner *Tuner) concurrent(uri string, needReq bool) error {
	var sucCount = new(int)
	var warnCount = new(int)
	var sucDetail = new(string)
	var failedDetail = new(string)
	wg := sync.WaitGroup{}
	for i, target := range tuner.Group {
		wg.Add(1)
		go func(i int, target Group, wg *sync.WaitGroup) {
			defer wg.Done()
			var request interface{}
			if needReq {
				request = target.MergedParam
			}

			result, allSuc := target.concurrentSuccess(uri, request)
			if allSuc {
				*sucCount++
				if result != "" {
					*warnCount++
				}
				*sucDetail += result
				return
			}

			*failedDetail += result
		}(i, target, &wg)
	}

	wg.Wait()
	retFailureInfo := strings.TrimSuffix(*failedDetail, "; ")
	switch uri {
	case "backup":
		tuner.backupFailure = retFailureInfo
	case "rollback":
		tuner.rollbackFailure = retFailureInfo
		if *sucDetail != "" {
			if *warnCount == len(tuner.Group) {
				tuner.rollbackDetail = fmt.Sprintf("All Targets No Need to Rollback")
				break
			}
			tuner.rollbackDetail = fmt.Sprintf("Partial success: %v No Need to Rollback", *sucDetail)
		}
	}

	if *sucCount != len(tuner.Group) {
		return fmt.Errorf("failure occur")
	}

	return nil
}
