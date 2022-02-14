package modules

import (
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
)

func (tuner *Tuner) loop() error {
	var err error
	var aheadStop bool
	for i := 1; i <= tuner.MAXIteration; i++ {
		tuner.Iteration = i
		// 1. acquire
		if aheadStop, err = tuner.acquire(); err != nil {
			return err
		}

		if aheadStop {
			break
		}

		// 2. apply
		if err = tuner.apply(); err != nil {
			return err
		}

		// 3. benchmark
		if err = tuner.benchmark(); err != nil {
			return err
		}

		// 4. feedback
		if err = tuner.feedback(); err != nil {
			return fmt.Errorf("feedback %vth configuration:%v", i, err)
		}

		// 5. analyse
		optimalRatioInfo := tuner.analyseResult()
		if optimalRatioInfo != "" {
			log.Infof(tuner.logName, "\tCurrent optimal iteration: %v\n", optimalRatioInfo)
		}

		if isInterrupted(tuner.logName) {
			log.Infof(tuner.logName, "Tuning interrupted after step%v, [loop tuning] round %v finish.", tuner.Step, i)
			return fmt.Errorf("tuning is interrupted")
		}
	}

	return nil
}

func (tuner *Tuner) benchmark() error {
	// get round of execution benchmark
	var round int
	if int(tuner.nextConfiguration.budget) != 0 {
		round = int(tuner.nextConfiguration.budget)
	} else {
		if tuner.isSensitize {
			round = config.KeenTune.Sensitize.BenchRound
		} else {
			round = config.KeenTune.ExecRound
		}
	}

	// execution benchmark
	var implyBenchResult string
	var err error
	tuner.Benchmark.LogName = tuner.logName
	tuner.benchScore, tuner.nextConfiguration.Score, implyBenchResult, err = tuner.Benchmark.RunBenchmark(round, &tuner.timeSpend.benchmark, tuner.Verbose)
	if err != nil {
		if err.Error() == "get benchmark is interrupted" {
			log.Infof(tuner.logName, "Tuning interrupted after step%v, [run benchmark] round %v stopped.", tuner.Step, tuner.Iteration)
			return fmt.Errorf("run benchmark interrupted")
		}
		return fmt.Errorf("tuning execute %vth benchmark err:%v", tuner.Iteration, err)
	}

	log.Infof(tuner.logName, "[Iteration %v] Benchmark result: %v", tuner.Iteration, implyBenchResult)
	tuner.TargetConfiguration[0].Score = tuner.nextConfiguration.Score
	// dump benchmark result of current tuning Iteration
	if config.KeenTune.DumpConf.ExecDump && !tuner.isSensitize {
		for index := range tuner.TargetConfiguration {
			targetID := index + 1
			tuner.TargetConfiguration[index].Score = tuner.nextConfiguration.Score
			tuner.TargetConfiguration[index].Dump(tuner.Name, fmt.Sprintf("_exec_%v_target_%v.json", tuner.Iteration, targetID))
		}
	}

	return nil
}

