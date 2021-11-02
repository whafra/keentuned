package modules

import (
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"keentune/daemon/common/utils/http"
	"encoding/json"
	"fmt"
	"time"
	"math"
)

// Tuner define a tuning job include Algorithm, Benchmark and Configurations
type Tuner struct {
	Name              string
	Scenario          string
	Algorithm         string // 使用的算法
	MAXIteration      int    // 最大执行轮次
	Iteration         int    // 当前轮次
	ClientHost        string
	BenchmarkHost     string
	StartTime         time.Time
	State             string
	Benchmark         Benchmark
	EnableParameters  []Parameter
	Configurations    []Configuration
	BaseConfiguration Configuration
	BestConfiguration Configuration
	nextConfiguration Configuration
	timeSpend         TimeSpend
	ParamConf         string
	Verbose           bool
	Step              int           // tuning process steps
	isSensitize       bool          // sensitive parameter identification mark
	Flag              string        // command flag, enum: "collect", "tuning"
	bestWeightedScore WeightedScore // current optimal weighted score details
}

type TimeSpend struct {
	init           time.Duration
	acquire        time.Duration
	apply          time.Duration
	send           time.Duration
	benchmark      time.Duration
	feedback       time.Duration
	end            time.Duration
	best           time.Duration
	detailInfo     string
}

type WeightedScore struct {
	Score    float32
	Info     string
	Round    int
}

// Tune : tuning main process
func (tuner *Tuner) Tune() {
	var err error

	defer func() {
		if err != nil {
			tuner.end()
		}
	}()

	if err = tuner.prepare(); err != nil {
		log.Errorf(log.ParamTune, "prepare for tuning err:%v", err)
		return
	}

	log.Infof(log.ParamTune, "\nStep%v. Start tuning, total iteration is %v.\n", tuner.IncreaseStep(), tuner.MAXIteration)

	if err = tuner.loop(); err != nil {
		log.Errorf(log.ParamTune, "loop tuning err:%v", err)
		return
	}

	if err = tuner.getBestConfiguration(); err != nil {
		log.Errorf(log.ParamTune, "get best configuration err:%v", err)
		return
	}

	if err = tuner.checkBestConfiguration(); err != nil {
		log.Errorf(log.ParamTune, "check best configuration err:%v", err)
		return
	}
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
	if isInterrupted() {
		log.Infof(logName, "Tuning interrupted after step%v, [acquire] round %v finish.", tuner.Step, tuner.Iteration)
		return false, fmt.Errorf("tuning is interrupted")
	}

	// check end loop ahead of time
	if acquiredInfo.Iteration < 0 {
		log.Warnf(logName, "%vth Tuning acquired round is less than zero, the tuning job will end ahead of time", tuner.Iteration)
		return true , nil
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

func (tuner *Tuner) apply(logName string) error{
	var err error
	tuner.nextConfiguration, err = tuner.nextConfiguration.Apply(&tuner.timeSpend.apply)
	if err != nil {
		if err.Error() == "get apply result is interrupted" {
			log.Infof(logName, "Tuning interrupted after step%v, [apply] round %v stopped.", tuner.Step, tuner.Iteration)
			return fmt.Errorf("run apply interrupted")
		}
		return fmt.Errorf("apply %vth configuration err:%v", tuner.Iteration, err)
	}

	if tuner.Verbose {
		log.Infof(logName, "[Iteration %v] Apply success, %v", tuner.Iteration, tuner.nextConfiguration.timeSpend.Desc)
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
	if _, tuner.nextConfiguration.Score, implyBenchResult, err = tuner.Benchmark.RunBenchmark(round, &tuner.timeSpend.benchmark, tuner.Verbose); err != nil {
		if err.Error() == "get benchmark is interrupted" {
			log.Infof(logName, "Tuning interrupted after step%v, [run benchmark] round %v stopped.", tuner.Step, tuner.Iteration)
			return fmt.Errorf("run benchmark interrupted")
		}
		return fmt.Errorf("tuning execute %vth benchmark err:%v", tuner.Iteration, err)
	}

	log.Infof(logName, "[Iteration %v] Benchmark result: %v", tuner.Iteration, implyBenchResult)

	// dump benchmark result of current tuning Iteration 
	if config.KeenTune.DumpConf.ExecDump && !tuner.isSensitize {
		tuner.nextConfiguration.Dump(tuner.Name, fmt.Sprintf("_exec_%v.json", tuner.Iteration))
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
		return fmt.Errorf("[getBestConfiguration] remote call  err:%v\n", err)
	}

	var bestConfig ReceivedConfigure
	err = json.Unmarshal(resp, &bestConfig)
	if err != nil {
		return fmt.Errorf("[getBestConfiguration] unmarshal best result err:%v\n", err)
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
	tuner.updateFeedbackConfig(&configuration)
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

func (tuner *Tuner) updateFeedbackConfig(configuration *Configuration) {
	for name, info := range tuner.BaseConfiguration.Score {
		score, ok := configuration.Score[name]
		if !ok {
			log.Warnf("", "feedback get [%v] from configure passed in  not exist", name)
			continue
		}
		score.Baseline = info.Value
		configuration.Score[name] = score
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

	initRatio := tuner.timeSpend.init.Seconds() * 100 / totalTime
	applyRatio := tuner.timeSpend.apply.Seconds() * 100 / totalTime
	acquireRatio := tuner.timeSpend.acquire.Seconds() * 100 / totalTime
	benchmarkRatio := tuner.timeSpend.benchmark.Seconds() * 100 / totalTime
	feedbackRatio := tuner.timeSpend.feedback.Seconds() * 100 / totalTime

	tuner.timeSpend.detailInfo = fmt.Sprintf("\n\t| %v\t| %v | %v \t| %v |", "Process", "Execution Count", "Total Time", "The Share of Total Time")
	tuner.timeSpend.detailInfo += formatRuntimeInfo("init", 1, tuner.timeSpend.init.Seconds(), initRatio)
	tuner.timeSpend.detailInfo += formatRuntimeInfo("apply", tuner.MAXIteration + 2, tuner.timeSpend.apply.Seconds(), applyRatio)
	tuner.timeSpend.detailInfo += formatRuntimeInfo("acquire", tuner.MAXIteration, tuner.timeSpend.acquire.Seconds(), acquireRatio)
	tuner.timeSpend.detailInfo += formatRuntimeInfo("benchmark", tuner.MAXIteration * config.KeenTune.ExecRound + config.KeenTune.BaseRound + config.KeenTune.AfterRound, tuner.timeSpend.benchmark.Seconds(), benchmarkRatio)
	tuner.timeSpend.detailInfo += formatRuntimeInfo("feedback", tuner.MAXIteration, tuner.timeSpend.feedback.Seconds(), feedbackRatio)
}

// prepare imply baseline operation, include init、apply baseline config、send file、base benchmark etc.
func (tuner *Tuner) prepare() error {
	emptyConfiguration, err := tuner.init()
	if err != nil {
		return fmt.Errorf("Step%v. tuning init failed, reason: %v", tuner.IncreaseStep(), err)
	}

	if isInterrupted() {
		log.Infof(log.ParamTune, "Tuning interrupted after step%v, [init] finish.", tuner.Step)
		return fmt.Errorf("tuning is interrupted")
	}

	log.Infof(log.ParamTune, "Step%v. AI Engine is ready.\n", tuner.IncreaseStep())
	tuner.BaseConfiguration, err = emptyConfiguration.Apply(&tuner.timeSpend.apply)
	if err != nil {		
		return fmt.Errorf("apply baseline configuration err:%v", err)
	}

	tuner.BaseConfiguration.Round = 0
	success, _, err := tuner.Benchmark.SendScript(&tuner.timeSpend.send)
	if err != nil || !success {
		return fmt.Errorf("send script file  result: %v err:%v", success, err)
	}

	log.Infof(log.ParamTune, "Step%v. Run benchmark as baseline:", tuner.IncreaseStep())

	var implyBenchResult string
	if _, tuner.BaseConfiguration.Score, implyBenchResult, err = tuner.Benchmark.RunBenchmark(config.KeenTune.BaseRound, &tuner.timeSpend.benchmark, tuner.Verbose); err != nil {
		if err.Error() == "get benchmark is interrupted" {
			log.Infof(log.ParamTune, "Tuning interrupted after step%v, [baseline benchmark] stopped.", tuner.Step)
			return fmt.Errorf("run benchmark interrupted")
		}
		return fmt.Errorf("tuning execute baseline benchmark err:%v", err)
	}

	log.Infof(log.ParamTune, "%v", implyBenchResult)
	if config.KeenTune.DumpConf.BaseDump {
		tuner.BaseConfiguration.Dump(tuner.Name, "_base.json")
	}

	if isInterrupted() {
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
		if err = tuner.feedback(tuner.nextConfiguration); err != nil {
			return fmt.Errorf("feedback %vth configuration err:%v", i, err)
		}

		// 5. analyse
		_ = tuner.analyseResult(tuner.nextConfiguration)
		if tuner.bestWeightedScore.Info != "" {
			log.Infof(logName, "   Current optimal iteration: [Iteration %v] %v\n", tuner.bestWeightedScore.Round, tuner.bestWeightedScore.Info)
		}

		if isInterrupted() {
			log.Infof(logName, "Tuning interrupted after step%v, [loop tuning] round %v finish.", tuner.Step, i)
			return fmt.Errorf("tuning is interrupted")
		}
	}

	return nil
}

func (tuner *Tuner) checkBestConfiguration() error {
	var implyBenchResult string
	var err error
	tuner.BestConfiguration, err = tuner.BestConfiguration.Apply(&tuner.timeSpend.apply)
	if err != nil {		
		log.Errorf(log.ParamTune, "apply best configuration err:%v\n", err)
		return err
	}

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
	log.Infof(log.ParamTune, "[BEST] Tuning improvment: %v\n", tuner.analyseResult(tuner.BestConfiguration))

	tuner.end()

	if tuner.Verbose {		
		log.Infof(log.ParamTune, "Time cost statistical information:%v", tuner.timeSpend.detailInfo)
	}

	return nil
}

func (tuner *Tuner) analyseResult(config Configuration) string {
	var increaseString string
	var weightedScore WeightedScore
	for name, info := range config.Score {
		base, ok := tuner.BaseConfiguration.Score[name]
		if !ok {
			log.Debugf("", "get baseline [%v] info not exist, please check the bench.json and the python file you specified whether matched", name)
			continue
		}

		score := utils.IncreaseRatio(info.Value, base.Value)
		weightedScore.Score += score*info.Weight
		if tuner.Verbose {
			if score < 0.0 && info.Negative || score > 0.0 && !info.Negative {
				increaseString += fmt.Sprintf("\n	[%v]\tImproved by %s;\t(baseline = %.3f)", name, utils.ColorString("Green", fmt.Sprintf("%.3f%%", math.Abs(float64(score)))), base.Value)

			} else {
				increaseString += fmt.Sprintf("\n	[%v]\tDeclined by %s;\t(baseline = %.3f)", name, utils.ColorString("Red", fmt.Sprintf("%.3f%%", math.Abs(float64(score)))), base.Value)				
			}
		}

		if !tuner.Verbose && info.Weight > 0.0 {
			if score < 0.0 && info.Negative || score > 0.0 && !info.Negative {				
				increaseString += fmt.Sprintf("\n	[%v]\tImproved by %s", name, utils.ColorString("Green", fmt.Sprintf("%.3f%%", math.Abs(float64(score)))))
				
			} else {				
				increaseString += fmt.Sprintf("\n	[%v]\tDeclined by %s", name, utils.ColorString("Red", fmt.Sprintf("%.3f%%", math.Abs(float64(score)))))
			}
		}
	}

	if tuner.bestWeightedScore.Score < weightedScore.Score {
		weightedScore.Info = increaseString
		weightedScore.Round = tuner.Iteration
		tuner.bestWeightedScore = weightedScore
	}

	return increaseString
}

// Collect Sensitive parameters
func (tuner *Tuner) Collect() {
	defer tuner.end()

	if err := tuner.initCollect(); err != nil {
		log.Errorf(log.SensitizeCollect, "initCollect failed, err:%v", err)
		return
	}

	log.Infof(log.SensitizeCollect, "\nStep%v. Collect init success.", tuner.IncreaseStep(1))
	log.Infof(log.SensitizeCollect, "\nStep%v. Start sensitization collection, total iteration is %v.\n", tuner.IncreaseStep(), tuner.MAXIteration)

	tuner.isSensitize = true
	if err := tuner.loop(); err != nil {
		log.Errorf(log.SensitizeCollect, "collect err:%v\n", err)
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

	if isInterrupted() {
		log.Infof(log.SensitizeCollect, "Collect interrupted after step%v, [init] finish.", tuner.Step)
		return fmt.Errorf("Collect is interrupted")
	}

	tuner.BaseConfiguration, err = emptyConfiguration.Apply(&tuner.timeSpend.apply)
	if err != nil {
		return fmt.Errorf("emptyConfiguration apply failed, err:%v", err)
	}

	success, _, err := tuner.Benchmark.SendScript(&tuner.timeSpend.send)
	if err != nil || !success {
		return fmt.Errorf("send script file  result: %v err:%v", success, err)
	}

	return nil
}

func formatRuntimeInfo(processName string, execCount int, totalTime, shareTime float64) string {
	return fmt.Sprintf("\n\t| %v  \t|  %v\t\t  |   %v \t|     %v  \t\t  |", processName, fmt.Sprintf("%v", execCount), fmt.Sprintf("%.3fs", totalTime), fmt.Sprintf("%.2f%%", shareTime))
}
