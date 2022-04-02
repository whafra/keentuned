package param

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	com "keentune/daemon/api/common"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	m "keentune/daemon/modules"
	"os"
	"sort"
	"strings"
	"time"
)

// TuneFlag tune options
type TuneFlag struct {
	ParamMap  string
	Name      string
	Round     int
	BenchConf string
	Verbose   bool
	Log       string
}

// Tune run param tune service
func (s *Service) Tune(flag TuneFlag, reply *string) error {
	if com.GetRunningTask() != "" {
		log.Errorf("", "Job %v is running, you can wait for it finishing or stop it.", com.GetRunningTask())
		return fmt.Errorf("Job %v is running, you can wait for finishing it or stop it.", com.GetRunningTask())
	}

	if err := com.HeartbeatCheck(); err != nil {
		return fmt.Errorf("check %v", err)
	}

	go runTuning(flag)
	return nil
}

func runTuning(flag TuneFlag) {
	com.SetRunningTask(com.JobTuning, flag.Name)
	log.ParamTune = "param tune" + ":" + flag.Log
	// create log file
	ioutil.WriteFile(flag.Log, []byte{}, os.ModePerm)
	defer func() {
		config.ProgramNeedExit <- true
		<-config.ServeFinish
		com.ClearTask()
	}()

	log.Infof(log.ParamTune, "Step1. Parameter auto tuning start, using algorithm = %v.\n", config.KeenTune.Algorithm)
	if err := TuningImpl(flag, "tuning"); err != nil {
		log.Errorf(log.ParamTune, "Param Tune failed, msg: %v", err)
		return
	}

}

func TuningImpl(flag TuneFlag, cmd string) error {
	benchInfo, err := GetBenchmarkInst(config.KeenTune.BenchConf)
	if err != nil {
		return err
	}

	tuner := &m.Tuner{
		MAXIteration: flag.Round,
		Name:         flag.Name,
		StartTime:    time.Now(),
		Verbose:      flag.Verbose,
		Step:         1,
		Flag:         cmd,
		Benchmark:    *benchInfo,
	}

	if cmd == "tuning" {
		tuner.Algorithm = config.KeenTune.Algorithm
		tuner.Tune()
		return nil
	}

	if cmd == "collect" {
		tuner.Algorithm = config.KeenTune.Sensitize.Algorithm
		tuner.Collect()
		return nil
	}

	return nil
}

func GetBenchmarkInst(benchFile string) (*m.Benchmark, error) {
	benchConf := config.GetBenchJsonPath(benchFile)
	if !file.IsPathExist(benchConf) {
		return nil, fmt.Errorf("Read BenchConf file [%v] err, file absolute path [%v] does not exist", benchFile, benchConf)
	}

	reqData, err := ioutil.ReadFile(benchConf)
	if err != nil {
		return nil, fmt.Errorf("Read bench conf file err: %v", err)
	}

	var bench map[string][]m.Benchmark
	if err = json.Unmarshal(reqData, &bench); err != nil {
		return nil, fmt.Errorf("Unmarshal bench conf file err: %v", err)
	}

	benchmarks := bench["benchmark"]
	if len(benchmarks) == 0 {
		return nil, fmt.Errorf("benchmark json is null")
	}

	tuneIP := strings.Split(config.KeenTune.Bench.DestIP, ",")
	if len(tuneIP) == 0 {
		return nil, fmt.Errorf("tune ip from keentuned.conf is empty")
	}

	inst := benchmarks[0]
	inst.Cmd = strings.Replace(strings.Replace(inst.Cmd, "{remote_script_path}", inst.FilePath, 1), "{target_ip}", tuneIP[0], 1)
//	inst.Host = fmt.Sprintf("%s:%s", config.KeenTune.SrcIPs, config.KeenTune.SrcPort)
	inst.SortedItems = sortBenchItemNames(inst.Items)
	return &inst, nil
}

func sortBenchItemNames(items map[string]m.ItemDetail) []string {
	var sortNames []string
	for name, _ := range items {
		sortNames = append(sortNames, name)
	}

	sort.Strings(sortNames)
	return sortNames
}

