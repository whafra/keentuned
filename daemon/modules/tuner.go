package modules

import (
	"encoding/json"
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"keentune/daemon/common/utils/http"
	"math"
	"strings"
	"time"
)

// Tuner define a tuning job include Algorithm, Benchmark and Configurations
type Tuner struct {
	Name                string
	Scenario            string
	Algorithm           string // 使用的算法
	MAXIteration        int    // 最大执行轮次
	Iteration           int    // 当前轮次
	TargetHost          []string
	BenchmarkHost       string
	StartTime           time.Time
	State               string
	Benchmark           Benchmark
	EnableParameters    []Parameter
	Configurations      []Configuration
	BaseConfiguration   []Configuration
	BestConfiguration   Configuration
	nextConfiguration   Configuration
	TargetConfiguration []Configuration
	timeSpend           TimeSpend
	ParamConf           string
	Verbose             bool
	Step                int         // tuning process steps
	isSensitize         bool        // sensitive parameter identification mark
	Flag                string      // command flag, enum: "collect", "tuning"
	bestItemsScore      []itemScore // current optimal score
}

type TimeSpend struct {
	init       time.Duration
	acquire    time.Duration
	apply      time.Duration
	send       time.Duration
	benchmark  time.Duration
	feedback   time.Duration
	end        time.Duration
	best       time.Duration
	detailInfo string
}

type itemScore struct {
	Score float32
	Info  string
	Round int
}

// Tune : tuning main process
func (tuner *Tuner) Tune() {
	var err error

	defer func() {
		if err != nil {
			tuner.end()
			parseTuningError(log.ParamTune, err)
		}
	}()

	if err = tuner.prepare(); err != nil {
		err = fmt.Errorf("prepare for tuning :%v", err)
		return
	}

	log.Infof(log.ParamTune, "\nStep%v. Start tuning, total iteration is %v.\n", tuner.IncreaseStep(), tuner.MAXIteration)

	if err = tuner.loop(); err != nil {
		err = fmt.Errorf("loop tuning err:%v", err)
		return
	}

	if err = tuner.getBestConfiguration(); err != nil {
		err = fmt.Errorf("get best configuration err:%v", err)
		return
	}

	if err = tuner.checkBestConfiguration(); err != nil {
		err = fmt.Errorf("check best configuration err:%v", err)
		return
	}
}

func parseTuningError(logName string, err error) {
	if err == nil {
		return
	}

	if strings.Contains(err.Error(), "apply configuration failed") {
		Rollback(logName)
	}

	if strings.Contains(err.Error(), "interrupted") {
		log.Infof(logName, "parameter optimization job abort!")
		return
	}
	log.Infof(logName, "%v", err)
}

/*acquire configuration from brain*/
func (tuner *Tuner) acquire(logName string) (bool, error) {
	// remote call and parse info
	start := time.Now()
	url := config.KeenTune.BrainIP + ":" + config.KeenTune.BrainPort + "/acquire"
	resp, err := http.RemoteCall("GET", url, nil)
	if err != nil {
		log.Errorf(logName, "[%vth] remote call acquire configuration err:%v", tuner.Iteration, err)
		return false, err
	}

	var acquiredInfo ReceivedConfigure
	if err = json.Unmarshal(resp, &acquiredInfo); err != nil {
		log.Errorf(logName, "[%vth] parse acquire unmarshal err:%v\n", tuner.Iteration, err)
		return false, err
	}

	// check interrupted
	if isInterrupted(logName) {
		log.Infof(logName, "Tuning interrupted after step%v, [acquire] round %v finish.", tuner.Step, tuner.Iteration)
		return false, fmt.Errorf("tuning is interrupted")
	}

	// check end loop ahead of time
	if acquiredInfo.Iteration < 0 {
		log.Warnf(logName, "%vth Tuning acquired round is less than zero, the tuning job will end ahead of time", tuner.Iteration)
		return true, nil
	}

	// time cost
	timeCost := utils.Runtime(start)
	tuner.timeSpend.acquire += timeCost.Count
	if tuner.Verbose {
		log.Infof(logName, "[Iteration %v] Acquire success, %v", tuner.Iteration, timeCost.Desc)
	}

	// assign to nextConfiguration
	tuner.nextConfiguration = Configuration{
		Parameters: acquiredInfo.Candidate,
		Round:      acquiredInfo.Iteration,
		budget:     acquiredInfo.Budget,
	}

	return false, nil
}

func (tuner *Tuner) apply(logName string) error {
	var err error
	var implyApplyResults string
	implyApplyResults, tuner.TargetConfiguration, err = tuner.nextConfiguration.Apply(&tuner.timeSpend.apply)
	if err != nil {
		return fmt.Errorf("%vth apply configuration failed:%v, details: %v", tuner.Iteration, err, implyApplyResults)
	}

	log.Debugf(logName, "step%v, [apply] round %v details: %v", tuner.Step, tuner.Iteration, implyApplyResults)

	if tuner.Verbose {
		var applyRuntimeInfo string
		for index, configuration := range tuner.TargetConfiguration {
			applyRuntimeInfo += fmt.Sprintf("\n\ttarget [%v] use time: %.3f s", index+1, configuration.timeSpend.Count.Seconds())
		}
		log.Infof(logName, "[Iteration %v] Apply success, details: %v", tuner.Iteration, applyRuntimeInfo)
	}

	return err
}

func (tuner *Tuner) benchmark(logName string) error {
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
	tuner.Benchmark.LogName = logName
	_, tuner.nextConfiguration.Score, implyBenchResult, err = tuner.Benchmark.RunBenchmark(round, &tuner.timeSpend.benchmark, tuner.Verbose)
	if err != nil {
		if err.Error() == "get benchmark is interrupted" {
			log.Infof(logName, "Tuning interrupted after step%v, [run benchmark] round %v stopped.", tuner.Step, tuner.Iteration)
			return fmt.Errorf("run benchmark interrupted")
		}
		return fmt.Errorf("tuning execute %vth benchmark err:%v", tuner.Iteration, err)
	}

	log.Infof(logName, "[Iteration %v] Benchmark result: %v", tuner.Iteration, implyBenchResult)
	tuner.TargetConfiguration[0].Score = tuner.nextConfiguration.Score
	// dump benchmark result of current tuning Iteration
	if config.KeenTune.DumpConf.ExecDump && !tuner.isSensitize {
		for index := range tuner.TargetConfiguration {
			targetID := index + 1
			tuner.TargetConfiguration[index].Score = tuner.nextConfiguration.Score
			tuner.TargetConfiguration[index].Dump(tuner.Name, fmt.Sprintf("_exec_%v_target_%v.json", tuner.Iteration, targetID))
		}
	}

	// add tuner Configurations
	tuner.Configurations = append(tuner.Configurations, tuner.nextConfiguration)
	return nil
}

func (tuner *Tuner) getBestConfiguration() error {
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

	// dump best configuration
	if config.KeenTune.DumpConf.BestDump {
		tuner.BestConfiguration.Dump(tuner.Name, "_best.json")
		log.Infof(log.ParamTune, "Step%v. Best configuration dump to [%v/parameter/%v/%v] successfully.\n", tuner.IncreaseStep(), config.KeenTune.DumpConf.DumpHome, tuner.Name, tuner.Name+"_best.json")
	}
	return nil
}

/*Feedback configuration with score to brain*/
func (tuner *Tuner) feedback(configuration Configuration) error {
	start := time.Now()
	tuner.updateFeedbackScore(&configuration.Score)
	feedbackMap := map[string]interface{}{
		"iteration": configuration.Round,
		"score":     configuration.Score,
	}
	err := http.ResponseSuccess("POST", config.KeenTune.BrainIP+":"+config.KeenTune.BrainPort+"/feedback", feedbackMap)
	if err != nil {
		return fmt.Errorf("[feedback] remote call feedback err:%v\n", err)
	}

	timeCost := utils.Runtime(start)
	tuner.timeSpend.feedback += timeCost.Count
	return nil
}

func (tuner *Tuner) updateFeedbackScore(scores *map[string]ItemDetail) {
	for name, info := range tuner.BaseConfiguration[0].Score {
		score, ok := (*scores)[name]
		if !ok {
			log.Warnf("", "feedback get [%v] from configure passed in  not exist", name)
			continue
		}
		score.Baseline = info.Value
		(*scores)[name] = score
	}
}

func (tuner *Tuner) init() (*Configuration, error) {
	start := time.Now()
	emptyConf, requireConf := generateInitParams(tuner.ParamConf)
	if emptyConf == nil || requireConf == nil {
		return nil, fmt.Errorf("read or assemble parameter failed")
	}

	requireConf["algorithm"] = tuner.Algorithm
	requireConf["iteration"] = tuner.MAXIteration
	requireConf["name"] = tuner.Name
	requireConf["type"] = tuner.Flag

	err := http.ResponseSuccess("POST", config.KeenTune.BrainIP+":"+config.KeenTune.BrainPort+"/init", requireConf)
	if err != nil {
		return nil, fmt.Errorf("remote call [init] failed: %v", err)
	}

	timeCost := utils.Runtime(start)
	tuner.timeSpend.init += timeCost.Count
	return emptyConf, nil
}

func (tuner *Tuner) end() {
	start := time.Now()
	http.RemoteCall("GET", config.KeenTune.BrainIP+":"+config.KeenTune.BrainPort+"/end", nil)
	timeCost := utils.Runtime(start)
	tuner.timeSpend.end += timeCost.Count

	totalTime := utils.Runtime(tuner.StartTime).Count.Seconds()

	if totalTime == 0.0 || !tuner.Verbose {
		return
	}

	tuner.setTimeCostToTableString(totalTime)
}

func (tuner *Tuner) setTimeCostToTableString(totalTime float64) {
	initRatio := fmt.Sprintf("%.2f%%", tuner.timeSpend.init.Seconds()*100/totalTime)
	applyRatio := fmt.Sprintf("%.2f%%", tuner.timeSpend.apply.Seconds()*100/totalTime)
	acquireRatio := fmt.Sprintf("%.2f%%", tuner.timeSpend.acquire.Seconds()*100/totalTime)
	benchmarkRatio := fmt.Sprintf("%.2f%%", tuner.timeSpend.benchmark.Seconds()*100/totalTime)
	feedbackRatio := fmt.Sprintf("%.2f%%", tuner.timeSpend.feedback.Seconds()*100/totalTime)

	var detailSlice [][]string
	header := []string{"Process", "Execution Count", "Total Time", "The Share of Total Time"}
	detailSlice = append(detailSlice, header)

	initTime := fmt.Sprintf("%.3fs", tuner.timeSpend.init.Seconds())
	detailSlice = append(detailSlice, []string{"init", "1", initTime, initRatio})

	applyRound := fmt.Sprint(tuner.MAXIteration + 2)
	applyTime := fmt.Sprintf("%.3fs", tuner.timeSpend.apply.Seconds())
	detailSlice = append(detailSlice, []string{"apply", applyRound, applyTime, applyRatio})

	maxRound := fmt.Sprint(tuner.MAXIteration)
	acquireTime := fmt.Sprintf("%.3fs", tuner.timeSpend.acquire.Seconds())
	detailSlice = append(detailSlice, []string{"acquire", maxRound, acquireTime, acquireRatio})

	benchRound := fmt.Sprint(tuner.MAXIteration*config.KeenTune.ExecRound + config.KeenTune.BaseRound + config.KeenTune.AfterRound)
	benchTime := fmt.Sprintf("%.3fs", tuner.timeSpend.benchmark.Seconds())
	detailSlice = append(detailSlice, []string{"benchmark", benchRound, benchTime, benchmarkRatio})

	feedbackTime := fmt.Sprintf("%.3fs", tuner.timeSpend.feedback.Seconds())
	detailSlice = append(detailSlice, []string{"feedback", maxRound, feedbackTime, feedbackRatio})

	tuner.timeSpend.detailInfo = utils.FormatInTable(detailSlice)
}

// prepare imply baseline operation, include init、apply baseline config、send file、base benchmark etc.
func (tuner *Tuner) prepare() error {
	emptyConfiguration, err := tuner.init()
	if err != nil {
		return fmt.Errorf("Step%v. tuning init failed, reason: %v", tuner.IncreaseStep(), err)
	}

	if isInterrupted(log.ParamTune) {
		log.Infof(log.ParamTune, "Tuning interrupted after step%v, [init] finish.", tuner.Step)
		return fmt.Errorf("tuning is interrupted")
	}

	log.Infof(log.ParamTune, "Step%v. AI Engine is ready.\n", tuner.IncreaseStep())
	var implyApplyResults string
	implyApplyResults, tuner.BaseConfiguration, err = emptyConfiguration.Apply(&tuner.timeSpend.apply)
	if err != nil {
		return fmt.Errorf("baseline apply configuration failed: %v, details: %v", err, implyApplyResults)
	}

	if err = backup(log.ParamTune, emptyConfiguration); err != nil {
		return err
	}

	log.Debugf(log.ParamTune, "Step%v. apply baseline configuration details: %v", tuner.Step+1, implyApplyResults)

	success, _, err := tuner.Benchmark.SendScript(&tuner.timeSpend.send)
	if err != nil || !success {
		return fmt.Errorf("send script file  result: %v, details:%v", success, err)
	}

	log.Infof(log.ParamTune, "Step%v. Run benchmark as baseline:", tuner.IncreaseStep())

	_, scoreResult, implyBenchResult, err := tuner.Benchmark.RunBenchmark(config.KeenTune.BaseRound, &tuner.timeSpend.benchmark, tuner.Verbose)
	if err != nil {
		if err.Error() == "get benchmark is interrupted" {
			log.Infof(log.ParamTune, "Tuning interrupted after step%v, [baseline benchmark] stopped.", tuner.Step)
			return fmt.Errorf("run benchmark interrupted")
		}
		return fmt.Errorf("tuning execute baseline benchmark:%v", err)
	}

	tuner.BaseConfiguration[0].Score = scoreResult

	log.Infof(log.ParamTune, "%v", implyBenchResult)
	if config.KeenTune.DumpConf.BaseDump {
		for index := range tuner.BaseConfiguration {
			targetID := index + 1
			tuner.BaseConfiguration[index].Round = 0
			tuner.BaseConfiguration[index].Score = scoreResult
			tuner.BaseConfiguration[index].Dump(tuner.Name, fmt.Sprintf("target_%v_base.json", targetID))
		}
	}

	if isInterrupted(log.ParamTune) {
		log.Infof(log.ParamTune, "Tuning interrupted after step%v, [baseline benchmark] finish.", tuner.Step)
		return fmt.Errorf("tuning is interrupted")
	}

	return nil
}

func (tuner *Tuner) loop() error {
	logName := log.ParamTune
	if tuner.Flag == "collect" {
		logName = log.SensitizeCollect
	}

	var err error
	var aheadStop bool
	tuner.bestItemsScore = make([]itemScore, len(tuner.Benchmark.SortedItems))

	for i := 1; i <= tuner.MAXIteration; i++ {
		tuner.Iteration = i
		// 1. acquire
		if aheadStop, err = tuner.acquire(logName); err != nil {
			return err
		}

		if aheadStop {
			break
		}

		// 2. apply
		if err = tuner.apply(logName); err != nil {
			return err
		}

		// 3. benchmark
		if err = tuner.benchmark(logName); err != nil {
			return err
		}

		// 4. feedback
		if err = tuner.feedback(tuner.TargetConfiguration[0]); err != nil {
			return fmt.Errorf("feedback %vth configuration:%v", i, err)
		}

		// 5. analyse
		optimalRatioInfo, _ := tuner.analyseResult(tuner.TargetConfiguration[0])
		if optimalRatioInfo != "" {
			log.Infof(logName, "\tCurrent optimal iteration: %v\n", optimalRatioInfo)
		}

		if isInterrupted(logName) {
			log.Infof(logName, "Tuning interrupted after step%v, [loop tuning] round %v finish.", tuner.Step, i)
			return fmt.Errorf("tuning is interrupted")
		}
	}

	return nil
}

func (tuner *Tuner) checkBestConfiguration() error {
	var implyBenchResult string
	implyApplyResults, bestConfiguration, err := tuner.BestConfiguration.Apply(&tuner.timeSpend.apply)
	if err != nil {
		log.Errorf(log.ParamTune, "best apply configuration failed:%v, details: %v", implyApplyResults)
		return err
	}
	log.Debugf(log.ParamTune, "Step%v. apply configuration details: %v", tuner.Step, implyApplyResults)

	tuner.BestConfiguration = bestConfiguration[0]

	log.Infof(log.ParamTune, "Step%v. Tuning is finished, checking benchmark score of best configuration.\n", tuner.IncreaseStep())

	if _, tuner.BestConfiguration.Score, implyBenchResult, err = tuner.Benchmark.RunBenchmark(config.KeenTune.AfterRound, &tuner.timeSpend.benchmark, tuner.Verbose); err != nil {
		if err.Error() == "get benchmark is interrupted" {
			log.Infof(log.ParamTune, "Tuning interrupted after step%v, [check best configuration benchmark] stopped.", tuner.Step)
			return fmt.Errorf("run benchmark interrupted")
		}
		log.Errorf(log.ParamTune, "tuning execute best benchmark err:%v\n", err)
		return err
	}

	log.Infof(log.ParamTune, "[BEST] Benchmark result: %v\n", implyBenchResult)

	_, currentRatioInfo := tuner.analyseResult(tuner.BestConfiguration)
	if currentRatioInfo != "" {
		log.Infof(log.ParamTune, "[BEST] Tuning improvement: %v\n", currentRatioInfo)
	}

	tuner.end()

	if tuner.Verbose {
		log.Infof(log.ParamTune, "Time cost statistical information:%v", tuner.timeSpend.detailInfo)
	}

	return nil
}

// analyseResult analyse benchmark score Result
func (tuner *Tuner) analyseResult(config Configuration) (string, string) {
	if tuner.isSensitize {
		return "", ""
	}

	var currentRatioInfo string
	for index, name := range tuner.Benchmark.SortedItems {
		info, ok := config.Score[name]
		if !ok {
			log.Warnf("", "%vth config [%v] info not exist", tuner.Iteration, name)
			continue
		}

		base, ok := tuner.BaseConfiguration[0].Score[name]
		if !ok {
			log.Warnf("", "get baseline [%v] info not exist, please check the bench.json and the python file you specified whether matched", name)
			continue
		}

		score, oneRatioInfo := getRatio(info, base, tuner.Verbose, name)
		if oneRatioInfo == "" {
			continue
		}

		currentRatioInfo += fmt.Sprintf("\n\t%v", oneRatioInfo)
		if tuner.Iteration == 1 {
			tuner.bestItemsScore[index].Info = fmt.Sprintf("\n\t[Iteration 1]\t%v", oneRatioInfo)
			tuner.bestItemsScore[index].Score = score
			continue
		}

		if (info.Negative && score < tuner.bestItemsScore[index].Score) || (!info.Negative && score > tuner.bestItemsScore[index].Score) {
			tuner.bestItemsScore[index].Info = fmt.Sprintf("\n\t[Iteration %v]\t%v", tuner.Iteration, oneRatioInfo)
			tuner.bestItemsScore[index].Score = score
		}
	}

	var bestRatioInfo string
	for _, item := range tuner.bestItemsScore {
		if item.Info != "" {
			bestRatioInfo += item.Info
		}
	}

	return bestRatioInfo, currentRatioInfo
}

func getRatio(info ItemDetail, base ItemDetail, verbose bool, name string) (float32, string) {
	score := utils.IncreaseRatio(info.Value, base.Value)
	if verbose {
		if (score < 0.0 && info.Negative) || (score > 0.0 && !info.Negative) {
			info := utils.ColorString("Green", fmt.Sprintf("%.3f%%", math.Abs(float64(score))))
			return score, fmt.Sprintf("[%v]\tImproved by %s;\t(baseline = %.3f)", name, info, base.Value)
		} else {
			info := utils.ColorString("Red", fmt.Sprintf("%.3f%%", math.Abs(float64(score))))
			return score, fmt.Sprintf("[%v]\tDeclined by %s;\t(baseline = %.3f)", name, info, base.Value)
		}
	}

	if !verbose && info.Weight > 0.0 {
		if (score < 0.0 && info.Negative) || (score > 0.0 && !info.Negative) {
			info := utils.ColorString("Green", fmt.Sprintf("%.3f%%", math.Abs(float64(score))))
			return score, fmt.Sprintf("[%v]\tImproved by %s", name, info)
		} else {
			info := utils.ColorString("Red", fmt.Sprintf("%.3f%%", math.Abs(float64(score))))
			return score, fmt.Sprintf("[%v]\tDeclined by %s", name, info)
		}
	}
	return score, ""
}

// Collect Sensitive parameters
func (tuner *Tuner) Collect() {
	var err error

	defer func() {
		tuner.end()
		parseTuningError(log.SensitizeCollect, err)
	}()

	if err = tuner.initCollect(); err != nil {
		err = fmt.Errorf("initCollect failed, err:%v", err)
		return
	}

	log.Infof(log.SensitizeCollect, "\nStep%v. Collect init success.", tuner.IncreaseStep(1))
	log.Infof(log.SensitizeCollect, "\nStep%v. Start sensitization collection, total iteration is %v.\n", tuner.IncreaseStep(), tuner.MAXIteration)

	tuner.isSensitize = true
	if err = tuner.loop(); err != nil {
		err = fmt.Errorf("collect err:%v\n", err)
		return
	}

	log.Infof(log.SensitizeCollect, "\nStep%v. Sensitization collection finished, you can get the result by the command \"keentune sensitize train\" (see more details: keentune sensitize train -h).", tuner.IncreaseStep())
}

func (tuner *Tuner) IncreaseStep(initVal ...int) int {
	if len(initVal) == 0 {
		tuner.Step++
		return tuner.Step
	}

	tuner.Step = initVal[0] + 1
	return tuner.Step
}

func (tuner *Tuner) initCollect() error {
	emptyConfiguration, err := tuner.init()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	if isInterrupted(log.SensitizeCollect) {
		log.Infof(log.SensitizeCollect, "Collect interrupted after step%v, [init] finish.", tuner.Step)
		return fmt.Errorf("Collect is interrupted")
	}

	var implyApplyResults string
	implyApplyResults, tuner.BaseConfiguration, err = emptyConfiguration.Apply(&tuner.timeSpend.apply)
	if err != nil {
		return fmt.Errorf("init apply configuration failed, details:%v", implyApplyResults)
	}

	log.Debugf(log.SensitizeCollect, "Step%v. apply configuration details: %v", tuner.Step+1, implyApplyResults)

	success, _, err := tuner.Benchmark.SendScript(&tuner.timeSpend.send)
	if err != nil || !success {
		return fmt.Errorf("send script file  result: %v err:%v", success, err)
	}

	return nil
}

func backup(logName string, conf *Configuration) error {
	requestInfo, err := conf.assembleApplyRequestMap()
	if err != nil {
		return fmt.Errorf("get backup request err: %v", err)
	}

	backupReq := utils.Parse2Map("data", requestInfo)
	if backupReq == nil {
		return fmt.Errorf("get backup request is null")
	}

	details, suc := Backup(logName, backupReq)
	if !suc {
		return fmt.Errorf("backup detail:\n%v", details)
	}

	return nil
}

