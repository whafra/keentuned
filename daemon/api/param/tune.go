package param

import (
	com "keentune/daemon/api/common"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	m "keentune/daemon/modules"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

// TuneFlag tune options
type TuneFlag struct {
	Name      string
	Round     int
	BenchConf string
	ParamConf string
	Verbose   bool
}

// Tune run param tune service
func (s *Service) Tune(flag TuneFlag, reply *string) error {
	go runTuning(flag)
	return nil
}

func runTuning(flag TuneFlag) {
	if com.SystemRun {
		log.Info(log.ParamTune, "An instance is running, please wait for it to finish and re-initiate the request.")
		return
	}

	com.SystemRun = true
	log.ClearCliLog(log.ParamTune)

	defer func() {
		config.ProgramNeedExit <- true
		<-config.ServeFinish
		com.SystemRun = false
	}()

	go com.HeartbeatCheck()

	if err := checkTuneNameRepeat(flag.Name); err != nil {
		log.Errorf(log.ParamTune, "Please change the tune name specified, err:%v", err)
		return
	}

	log.Infof(log.ParamTune, "\nStep1. Parameter auto tuning start, using algorithm = %v.\n", config.KeenTune.Algorithm)
	if err := TuningImpl(flag, "tuning"); err != nil {
		return
	}

}

func TuningImpl(flag TuneFlag, cmd string) error {
	start := time.Now()
	reqData, err := ioutil.ReadFile(config.KeenTune.Home + flag.BenchConf)
	if err != nil {
		log.Errorf(log.ParamTune, "readfile err:%v\n", err)
		return err
	}

	var bench map[string][]m.Benchmark
	if err = json.Unmarshal(reqData, &bench); err != nil {
		log.Errorf(log.ParamTune, "unmarshal err:%v\n", err)
		return err
	}

	tuneIP := strings.Split(config.KeenTune.TargetIP, ",")
	for index, host := range tuneIP {
		tuner := &m.Tuner{
			MAXIteration:  flag.Round,
			ClientHost:    host,
			BenchmarkHost: host,
			Name:          flag.Name,
			StartTime:     start,
			ParamConf:     flag.ParamConf,
			Verbose:       flag.Verbose,
			Step:          1,
			Flag:          cmd,
		}

		benchInfo := getBenchmark(bench, index)
		if benchInfo == nil {
			return fmt.Errorf("benchmark json is null")
		}

		tuner.Benchmark = *benchInfo

		tuner.Benchmark.Cmd = strings.Replace(strings.Replace(tuner.Benchmark.Cmd, "{remote_script_path}", tuner.Benchmark.FilePath, 1), "{target_ip}", config.KeenTune.TargetIP, 1)
		if cmd == "tuning" {
			tuner.Algorithm = config.KeenTune.Algorithm
			tuner.Loop()
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

	if len(benchmarks)-1 < index && len(benchmarks) >= 1 {
		return &benchmarks[0]
	}

	return &benchmarks[index]
}

func checkTuneNameRepeat(name string) error {
	if com.IsDataNameUsed(name) {
		return fmt.Errorf("The name [%v] is in use", name)
	}

	return nil
}
