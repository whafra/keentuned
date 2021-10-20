package param

import (
	com "keentune/daemon/api/common"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	m "keentune/daemon/modules"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"
	"os"
)

// TuneFlag tune options
type TuneFlag struct {
	Name      string
	Round     int
	BenchConf string
	ParamConf string
	Verbose   bool
	Log       string
}

// Tune run param tune service
func (s *Service) Tune(flag TuneFlag, reply *string) error {
	if com.SystemRun {
		log.Errorf("", "An instance is running. You can wait the process finish or run \"keentune %v stop\" and try a new job again, if you want give up the old job.", com.GetRunningTask())
		return fmt.Errorf("Tuning failed, an instance is running. You can wait the process finish or run \"keentune %v stop\" and try a new job again, if you want give up the old job.", com.GetRunningTask())
	}

	go runTuning(flag)
	return nil
}

func runTuning(flag TuneFlag) {
	com.SystemRun = true
	com.IsTuning = true
	log.ParamTune =  "param tune" + ":" + flag.Log
	// create log file
	ioutil.WriteFile(flag.Log, []byte{}, os.ModePerm)
	defer func() {
		log.ClearCliLog(log.ParamTune)
		config.ProgramNeedExit <- true
		<-config.ServeFinish
		com.SystemRun = false
		com.IsTuning = false
	}()

	go com.HeartbeatCheck()

	if isTuneNameRepeat(flag.Name) {
		log.Errorf(log.ParamTune, "The specified name [%v] already exists. Please delete the original job by [keentune param delete --job %v] or specify a new name and try again", flag.Name, flag.Name)
		return
	}

	log.Infof(log.ParamTune, "Step1. Parameter auto tuning start, using algorithm = %v.\n", config.KeenTune.Algorithm)
	if err := TuningImpl(flag, "tuning"); err != nil {
		log.Errorf(log.ParamTune, "Param Tune failed, msg:[%v]", err)
		return
	}

}

func TuningImpl(flag TuneFlag, cmd string) error {
	paramConf := com.GetAbsolutePath(flag.ParamConf, "parameter", ".json", "_best.json")
	if !file.IsPathExist(paramConf) {
		return fmt.Errorf("Read ParamConf file [%v] err: file absolute path [%v] does not exist", flag.ParamConf, paramConf)
	}
	start := time.Now()
	reqData, err := ioutil.ReadFile(GetBenchJsonPath(flag.BenchConf))
	if err != nil {
		return fmt.Errorf("Read BenchConf file %v err:%v", flag.BenchConf, err)
	}

	var bench map[string][]m.Benchmark
	if err = json.Unmarshal(reqData, &bench); err != nil {
		return fmt.Errorf("Unmarshal BenchConf err:%v", err)
	}

	tuneIP := strings.Split(config.KeenTune.TargetIP, ",")
	for index, host := range tuneIP {
		tuner := &m.Tuner{
			MAXIteration:  flag.Round,
			ClientHost:    host,
			Name:          flag.Name,
			StartTime:     start,
			ParamConf:     paramConf,
			Verbose:       flag.Verbose,
			Step:          1,
			Flag:          cmd,
		}

		benchInfo := getBenchmark(bench, index)
		if benchInfo == nil {
			return fmt.Errorf("Benchmark json is null")
		}

		tuner.Benchmark = *benchInfo

		tuner.Benchmark.Cmd = strings.Replace(strings.Replace(tuner.Benchmark.Cmd, "{remote_script_path}", tuner.Benchmark.FilePath, 1), "{target_ip}", host, 1)

		if cmd == "tuning" {
			tuner.Algorithm = config.KeenTune.Algorithm
			tuner.Tune()
			continue
		}

		if cmd == "collect" {
			tuner.Algorithm = config.KeenTune.Sensitize.Algorithm
			tuner.Collect()
			continue
		}
	}

	return nil
}

func getBenchmark(bench map[string][]m.Benchmark, index int) *m.Benchmark {
	benchmarks := bench["benchmark"]
	if len(benchmarks) == 0 {
		return nil
	}

	var benchmark *m.Benchmark
	if len(benchmarks)-1 < index && len(benchmarks) >= 1 {
		benchmark = &benchmarks[0]
	}else {
		benchmark = &benchmarks[index]
	}
	
	benchmark.Host = fmt.Sprintf("%s:%s", config.KeenTune.BenchIP, config.KeenTune.BenchPort)

	return benchmark
}

func isTuneNameRepeat(name string) bool {
	tuneList, err := file.WalkFilePath(m.GetTuningWorkPath("") + "/", "", true, "/generate/")
	if err != nil {
		return false
	}

	for _, has:= range tuneList {
		if has == name {
			return true
		}
	}

	return false
}

func GetBenchJsonPath(fileName string) string {
	if string(fileName[0]) == "/" || fileName == "" {		
		return fileName
	}

	parts :=strings.Split(fileName, "/")
    if len(parts) == 1 {
		benchPath, err := file.GetWalkPath(m.GetBenchHomePath(), fileName)
		if err != nil {
			return fileName
		}

		return benchPath
	}
	
	return fmt.Sprintf("%v/%v", m.GetBenchHomePath(), strings.TrimPrefix(fileName, "benchmark/"))
}
