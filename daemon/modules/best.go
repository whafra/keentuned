package modules

import (
	"encoding/json"
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"keentune/daemon/common/utils/http"
	"time"
)

func (tuner *Tuner) getBestConfiguration() error {
	err := tuner.requestBest()
	if err != nil {
		return err
	}

	// dump best configuration
	if config.KeenTune.DumpConf.BestDump {
		tuner.BestConfiguration.Dump(tuner.Name, "_best.json")
		log.Infof(log.ParamTune, "Step%v. Best configuration dump to [%v/parameter/%v/%v] successfully.\n", tuner.IncreaseStep(), config.KeenTune.DumpConf.DumpHome, tuner.Name, tuner.Name+"_best.json")
	}
	return nil
}

func (tuner *Tuner) requestBest() error {
	// get best configuration
	start := time.Now()
	url := config.KeenTune.BrainIP + ":" + config.KeenTune.BrainPort + "/best"
	resp, err := http.RemoteCall("GET", url, nil)
	if err != nil {
		return fmt.Errorf("remote call: %v\n", err)
	}

	var bestConfig ReceivedConfigure
	err = json.Unmarshal(resp, &bestConfig)
	if err != nil {
		return fmt.Errorf("unmarshal best config: %v\n", err)
	}

	tuner.BestConfiguration.Round = bestConfig.Iteration
	tuner.BestConfiguration.Parameters = bestConfig.Candidate
	tuner.BestConfiguration.Score = bestConfig.Score

	// time cost
	timeCost := utils.Runtime(start)
	tuner.timeSpend.best += timeCost.Count
	return nil
}

func (tuner *Tuner) checkBestConfiguration() error {
	var implyBenchResult string
	implyApplyResults, bestConfiguration, err := tuner.BestConfiguration.Apply(&tuner.timeSpend.apply, false)
	if err != nil {
		log.Errorf(log.ParamTune, "best apply configuration failed:%v, details: %v", implyApplyResults)
		return err
	}
	log.Debugf(log.ParamTune, "Step%v. apply configuration details: %v", tuner.Step, implyApplyResults)

	tuner.BestConfiguration = bestConfiguration[0]

	log.Infof(log.ParamTune, "Step%v. Tuning is finished, checking benchmark score of best configuration.\n", tuner.IncreaseStep())

	if tuner.benchScore, _, implyBenchResult, err = tuner.Benchmark.RunBenchmark(config.KeenTune.AfterRound, &tuner.timeSpend.benchmark, tuner.Verbose); err != nil {
		if err.Error() == "get benchmark is interrupted" {
			log.Infof(log.ParamTune, "Tuning interrupted after step%v, [check best configuration benchmark] stopped.", tuner.Step)
			return fmt.Errorf("run benchmark interrupted")
		}
		log.Errorf(log.ParamTune, "tuning execute best benchmark err:%v\n", err)
		return err
	}

	log.Infof(log.ParamTune, "[BEST] Benchmark result: %v\n", implyBenchResult)

	currentRatioInfo := tuner.analyseBestResult()
	if currentRatioInfo != "" {
		log.Infof(log.ParamTune, "[BEST] Tuning improvement: %v\n", currentRatioInfo)
	}

	tuner.end()

	if tuner.Verbose {
		log.Infof(log.ParamTune, "Time cost statistical information:%v", tuner.timeSpend.detailInfo)
	}

	return nil
}

