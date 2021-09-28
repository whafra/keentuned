package modules

import (
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"keentune/daemon/common/utils/http"
	"encoding/json"
	"fmt"
	"time"
)

// Tuner define a tuning job include Algorithm, Benchmark and Configurations
type Tuner struct {
	Name              string
	Scenario          string
	Algorithm         string //使用的算法
	MAXIteration      int    // 最大执行轮次
	Iteration         int    //当前轮次
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
	Step              int
	isSensitize       bool
	Flag              string //  enum: "collect", "tuning"
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

// Loop : tuning main loop
func (tuner *Tuner) Loop() {
	var err error

	defer func() {
		if err != nil {
			tuner.end()
		}
	}()

	if err = tuner.prepare(); err != nil {
		return
	}

	log.Infof(log.ParamTune, "\nStep%v. Start tuning, total iteration is %v.\n", tuner.IncreaseStep(), tuner.MAXIteration)
	if err = tuner.imply(); err != nil {
		log.Errorf(log.ParamTune, "tuning err:%v\n", err)
		return
	}

	if err = tuner.getBestConfiguration(); err != nil {
		log.Errorf(log.ParamTune, "get best configuration err:%v\n", err)
		return
	}

	if config.KeenTune.DumpConf.BestDump {
		tuner.BestConfiguration.Dump(tuner.Name, "_best.json")
		log.Infof(log.ParamTune, "Step%v. Best configuration dump to %v/parameter/%v/%v\n", tuner.IncreaseStep(), config.KeenTune.DumpConf.DumpHome, tuner.Name, tuner.Name+"_best.json")
	}

	if err = tuner.checkBestConfiguration(); err != nil {
		log.Errorf(log.ParamTune, "check best configuration err:%v\n", err)
		return
	}
}

/*acquire configuration from brain*/
func (tuner *Tuner) acquire() (Configuration, string, error) {
	start := time.Now()
	url := config.KeenTune.BrainIP + ":" + config.KeenTune.BrainPort + "/acquire"
	resp, err := http.RemoteCall("GET", url, nil)
	if err != nil {
		return Configuration{}, "", err
	}

	var acquiredInfo ReceivedConfigure
	if err = json.Unmarshal(resp, &acquiredInfo); err != nil {
		log.Errorf(log.ParamTune, "parse acquire unmarshal err:%v\n", err)
		return Configuration{}, "", err
	}

	retConfig := Configuration{
		Parameters: acquiredInfo.Candidate,
		Round:      acquiredInfo.Iteration,
		budget:     acquiredInfo.Budget,
	}

	timeCost := utils.Runtime(start)
	tuner.timeSpend.acquire += timeCost.Count

	return retConfig, timeCost.Desc, nil
}

func (tuner *Tuner) getBestConfiguration() error {
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

	timeCost := utils.Runtime(start)
	tuner.timeSpend.best += timeCost.Count
	return nil
}

/*Feedback configuration with score to brain*/
func (tuner *Tuner) feedback(configuration Configuration) error {
	start := time.Now()
	tuner.UpdateFeedbackConfig(&configuration)
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

func (tuner *Tuner) UpdateFeedbackConfig(configuration *Configuration) {
	for name, info := range tuner.BaseConfiguration.Score {
		score, ok := configuration.Score[name]
		if !ok {
			log.Warnf(log.ParamTune, "feedback get [%v] from configure passed in  not exist", name)
			continue
		}
		score.Baseline = info.Value
		configuration.Score[name] = score
	}
}

/*Get sensibility of parameter from brain*/
func (tuner *Tuner) sensibility() {}

func (tuner *Tuner) init() *Configuration {
	start := time.Now()
	emptyConf, requireConf := generateInitParams(tuner.ParamConf)
	if emptyConf == nil || requireConf == nil {
		log.Errorf(log.ParamTune, "emptyConf [%+v], requireConf  [%+v]\n", emptyConf, requireConf)
		return nil
	}

	requireConf["algorithm"] = tuner.Algorithm
	requireConf["iteration"] = tuner.MAXIteration
	requireConf["name"] = tuner.Name
	requireConf["type"] = tuner.Flag

	err := http.ResponseSuccess("POST", config.KeenTune.BrainIP+":"+config.KeenTune.BrainPort+"/init", requireConf)
	if err != nil {
		log.Errorf(log.ParamTune, "remote call [init] err:%v\n", err)
		return nil
	}

	timeCost := utils.Runtime(start)
	tuner.timeSpend.init += timeCost.Count
	return emptyConf
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

	tuner.timeSpend.detailInfo = fmt.Sprintf("%s,%s,%s,%s;", "Process", "Execution Count", "Total Time", "The Share of Total Time")
	tuner.timeSpend.detailInfo += formatRuntimeInfo("init", 1, tuner.timeSpend.init.Seconds(), initRatio)
	tuner.timeSpend.detailInfo += formatRuntimeInfo("apply", tuner.MAXIteration + 2, tuner.timeSpend.apply.Seconds(), applyRatio)
	tuner.timeSpend.detailInfo += formatRuntimeInfo("acquire", tuner.MAXIteration, tuner.timeSpend.acquire.Seconds(), acquireRatio)
	tuner.timeSpend.detailInfo += formatRuntimeInfo("benchmark", tuner.MAXIteration * config.KeenTune.ExecRound + config.KeenTune.BaseRound + config.KeenTune.AfterRound, tuner.timeSpend.benchmark.Seconds(), benchmarkRatio)
	tuner.timeSpend.detailInfo += formatRuntimeInfo("feedback", tuner.MAXIteration, tuner.timeSpend.feedback.Seconds(), feedbackRatio)
}

// prepare imply baseline operation, include init、apply baseline config、send file、base benchmark etc.
func (tuner *Tuner) prepare() error {
	emptyConfiguration := tuner.init()
	if emptyConfiguration == nil {
		log.Errorf(log.ParamTune, "Step%v. tuning init failed, because configuration is nil", tuner.IncreaseStep())
		return fmt.Errorf("init configuration failed")
	}

	if isInterrupted() {
		log.Infof(log.ParamTune, "Tuning interrupted after step%v, [init] finish.", tuner.Step)
		return fmt.Errorf("tuning is interrupted")
	}

	log.Infof(log.ParamTune, "\nStep%v. AI Engine is ready.\n", tuner.IncreaseStep())
	var err error
	tuner.BaseConfiguration, err = emptyConfiguration.Apply(&tuner.timeSpend.apply)
	if err != nil {		
		log.Errorf(log.ParamTune, "apply baseline configuration err:%v", err)
		return err
	}

	tuner.BaseConfiguration.Round = 0
	success, _, err := tuner.Benchmark.sendScript(&tuner.timeSpend.send)
	if err != nil || !success {
		log.Errorf(log.ParamTune, "send script file  result: %v err:%v", success, err)
		return fmt.Errorf("send file failed")
	}

	log.Infof(log.ParamTune, "Step%v. Run benchmark as baseline:", tuner.IncreaseStep())

	var implyBenchResult string
	if tuner.BaseConfiguration.Score, implyBenchResult, err = tuner.Benchmark.RunBenchmark(config.KeenTune.BaseRound, &tuner.timeSpend.benchmark, tuner.Verbose); err != nil {
		if err.Error() == "get benchmark is interrupted" {
			log.Infof(log.ParamTune, "Tuning interrupted after step%v, [baseline benchmark] stopped.", tuner.Step)
			return fmt.Errorf("run benchmark interrupted")
		}
		log.Errorf(log.ParamTune, "tuning execute baseline benchmark err:%v\n", err)
		return fmt.Errorf("run benchmark failed")
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

func (tuner *Tuner) imply() error {
	logName := log.ParamTune 
	if tuner.Flag == "collect" {
		logName = log.SensitizeCollect
	}
	var implyBenchResult, implyAcquireResult string
	var err error
	for i := 1; i <= tuner.MAXIteration; i++ {
		// 1. Acquire
		if tuner.nextConfiguration, implyAcquireResult, err = tuner.acquire(); err != nil {
			return fmt.Errorf("------ acquire %vth configuration err:%v", i, err)
		}

		if isInterrupted() {
			log.Infof(logName, "Tuning interrupted after step%v, [acquire] round %v finish.", tuner.Step, i)
			return fmt.Errorf("tuning is interrupted")
		}

		if tuner.nextConfiguration.Round < 0 {
			log.Warnf(logName, "%vth Tuning acquired round is less than zero, the tuning task will end ahead of time", i)
			break
		}

		if tuner.Verbose {
			log.Infof(logName, "[Iteration %v] Acquire success, %v", i, implyAcquireResult)
		}

		// 2. Apply
		tuner.nextConfiguration, err = tuner.nextConfiguration.Apply(&tuner.timeSpend.apply)
		if err != nil {
			if err.Error() == "get apply result is interrupted" {
				log.Infof(logName, "Tuning interrupted after step%v, [apply] round %v stopped.", tuner.Step, i)
				return fmt.Errorf("run apply interrupted")
			}
			return fmt.Errorf("apply %vth configuration err:%v", i, err)
		}

		if tuner.Verbose {
			log.Infof(logName, "[Iteration %v] Apply success, %v", i, tuner.nextConfiguration.timeSpend.Desc)
		}

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

		// 3. Benchmark
		if tuner.nextConfiguration.Score, implyBenchResult, err = tuner.Benchmark.RunBenchmark(round, &tuner.timeSpend.benchmark, tuner.Verbose); err != nil {
			if err.Error() == "get benchmark is interrupted" {
				log.Infof(logName, "Tuning interrupted after step%v, [run benchmark] round %v stopped.", tuner.Step, i)
				return fmt.Errorf("run benchmark interrupted")
			}
			return fmt.Errorf("tuning execute %vth benchmark err:%v", i, err)
		}

		log.Infof(logName, "[Iteration %v] Benchmark result: %v", i, implyBenchResult)

		if config.KeenTune.DumpConf.ExecDump && !tuner.isSensitize {
			tuner.nextConfiguration.Dump(tuner.Name, fmt.Sprintf("_exec_%v.json", i))
		}

		tuner.Configurations = append(tuner.Configurations, tuner.nextConfiguration)

		// 4. feedback
		if err = tuner.feedback(tuner.nextConfiguration); err != nil {
			return fmt.Errorf("feedback %vth configuration err:%v", i, err)
		}

		improvementString := tuner.analyseResult(tuner.nextConfiguration)
		if improvementString != "" {
			log.Infof(logName, "[Iteration %v] Tuning improvment: %v\n", i, improvementString)
		}

		if isInterrupted() {
			log.Infof(logName, "Tuning interrupted after step%v, [imply tuning] round %v finish.", tuner.Step, i)
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

	if tuner.BestConfiguration.Score, implyBenchResult, err = tuner.Benchmark.RunBenchmark(config.KeenTune.AfterRound, &tuner.timeSpend.benchmark, tuner.Verbose); err != nil {
		if err.Error() == "get benchmark is interrupted" {
			log.Infof(log.ParamTune, "Tuning interrupted after step%v, [check best exec benchmark] stopped.", tuner.Step)
			return fmt.Errorf("run benchmark interrupted")
		}
		log.Errorf(log.ParamTune, "tuning execute best benchmark err:%v\n", err)
		return err
	}

	log.Infof(log.ParamTune, "[BEST] Benchmark result: %v\n", implyBenchResult)
	log.Infof(log.ParamTune, "[BEST] Tuning improvment: %v\n", tuner.analyseResult(tuner.BestConfiguration))

	tuner.end()

	if tuner.Verbose {		
		log.Infof(log.ParamTune, "Time cost statistical information details displayed in the terminal.")
		log.Infof(log.ParamTune, "%s show table end.", tuner.timeSpend.detailInfo)
	}

	return nil
}

func (tuner *Tuner) analyseResult(config Configuration) string {
	var increaseString string
	for name, info := range config.Score {
		base, ok := tuner.BaseConfiguration.Score[name]
		if !ok {
			log.Debugf(log.ParamTune, "get baseline [%v] info not exist, please check the bench.json and the python file you specified whether matched", name)
			continue
		}

		score := utils.IncreaseRatio(info.Value, base.Value)
		if tuner.Verbose {
			if score < 0.0 {
				if info.Negative {
					increaseString += fmt.Sprintf("\n	[%v]\tImproved by %s;\t(baseline = %.3f)", name, utils.ColorString("Green", fmt.Sprintf("%.3f%%", -score)), base.Value)
				} else {
					increaseString += fmt.Sprintf("\n	[%v]\tDeclined by %s;\t(baseline = %.3f)", name, utils.ColorString("Red", fmt.Sprintf("%.3f%%", -score)), base.Value)
				}
			} else {
				if info.Negative {
					increaseString += fmt.Sprintf("\n	[%v]\tDeclined by %s;\t(baseline = %.3f)", name, utils.ColorString("Red", fmt.Sprintf("%.3f%%", score)), base.Value)
				} else {
					increaseString += fmt.Sprintf("\n	[%v]\tImproved by %s;\t(baseline = %.3f)", name, utils.ColorString("Green", fmt.Sprintf("%.3f%%", score)), base.Value)
				}
			}
		}

		if !tuner.Verbose && info.Weight > 0.0 {
			if score < 0.0 {
				if info.Negative {
					increaseString += fmt.Sprintf("\n	[%v]\tImproved by %s", name, utils.ColorString("Green", fmt.Sprintf("%.3f%%", -score)))
				} else {
					increaseString += fmt.Sprintf("\n	[%v]\tDeclined by %s", name, utils.ColorString("Red", fmt.Sprintf("%.3f%%", -score)))
				}
			} else {
				if info.Negative {
					increaseString += fmt.Sprintf("\n	[%v]\tDeclined by %s", name, utils.ColorString("Red", fmt.Sprintf("%.3f%%", score)))
				} else {
					increaseString += fmt.Sprintf("\n	[%v]\tImproved by %s", name, utils.ColorString("Green", fmt.Sprintf("%.3f%%", score)))
				}
			}
		}
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
	if err := tuner.imply(); err != nil {
		log.Errorf(log.SensitizeCollect, "collect err:%v\n", err)
		return
	}

	log.Infof(log.SensitizeCollect, "\nStep%v. Sensitization collection finished, you can get the result by the command [keentune sensitize train] (keentune sensitize train -h for help).", tuner.IncreaseStep())
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
	emptyConfiguration := tuner.init()
	if emptyConfiguration == nil {
		return fmt.Errorf("init failed, because configuration is nil")
	}

	if isInterrupted() {
		log.Infof(log.SensitizeCollect, "Collect interrupted after step%v, [init] finish.", tuner.Step)
		return fmt.Errorf("Collect is interrupted")
	}

	var err error
	tuner.BaseConfiguration, err = emptyConfiguration.Apply(&tuner.timeSpend.apply)
	if err != nil {		
		return fmt.Errorf("emptyConfiguration apply failed, err:%v", err)
	}

	success, _, err := tuner.Benchmark.sendScript(&tuner.timeSpend.send)
	if err != nil || !success {
		return fmt.Errorf("send script file  result: %v err:%v", success, err)
	}

	return nil
}

func formatRuntimeInfo(processName string, execCount int, totalTime, shareTime float64) string {
	return fmt.Sprintf("%s,%s,%s,%s;", processName, fmt.Sprintf("%v", execCount), fmt.Sprintf("%.3f", totalTime), fmt.Sprintf("%.2f%%", shareTime))
}
