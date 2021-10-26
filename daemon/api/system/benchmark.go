package system

import (
	"keentune/daemon/common/log"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	com "keentune/daemon/api/common"
	"keentune/daemon/api/param"
	m "keentune/daemon/modules"
	"time"
	"strings"
	"io/ioutil"
	"encoding/json"
	"fmt"
)

type Service struct {}

type BenchmarkFlag struct {
	Round     int
	BenchConf string
	Name      string
} 

func (s *Service) Benchmark(flag BenchmarkFlag, reply *string) error {
	if com.SystemRun {
		return fmt.Errorf("An tuning instance is running, please wait for it to finish or stop the tuning task and retry.")
	}

	com.SystemRun = true

	defer func() {
		*reply = log.ClientLogMap[log.Benchmark]
		log.ClearCliLog(log.Benchmark)
		com.SystemRun = false
	}()

	inst, err:=getBenchmarkInst(flag.BenchConf)
	if err!=nil {
		return err
	}

	var step int
	step++
	log.Infof(log.Benchmark, "Step%v. Get benchmark instance successfully.\n", step)

	var sendTimeSpend, benchTimeSpend time.Duration
	success, _, err := inst.SendScript(&sendTimeSpend)
	if err != nil || !success {
		log.Errorf(log.Benchmark, "send script file  result: %v err:%v", success, err)
		return fmt.Errorf("send file failed")
	}

	step++
	log.Infof(log.Benchmark, "Step%v. Send benchmark script successfully.\n", step)

	step++
	log.Infof(log.Benchmark, "Step%v. Run [%v] round benchmark ...\n", step, flag.Round)

	var score map[string][]float32
	var benchmarkResult string
	scores := make(map[string][]float32)

	for i := 1; i <= flag.Round; i++ {
		if score, _, benchmarkResult, err = inst.RunBenchmark(1, &benchTimeSpend, false); err != nil {
			if err.Error() == "get benchmark is interrupted" {
				return fmt.Errorf("run [%v] round benchmark positive stopped", i)
			}

			log.Errorf(log.Benchmark, "Run Benchmark err:%v\n", err)
			return fmt.Errorf("Run Benchmark err:%v", err)
		}

		log.Infof(log.Benchmark, "[Iteration %v] Benchmark result:%v", i, strings.Replace(benchmarkResult, "average ", "", 1))

		for key, value := range score {
			scores[key]= append(scores[key], value...)
		}
	}	

	if err = file.Save2CSV(m.GetDumpCSVPath(), flag.Name + ".csv", scores); err!=nil {
		log.Warnf(log.ParamTune, "Save  Baseline benchmark  to file %v failed, reason:[%v]", flag.Name, err)
	}

	step++
	log.Infof(log.Benchmark, "\nStep%v. Benchmark result dump to %v susccessfully.", step, fmt.Sprintf("%v/%v.csv", m.GetDumpCSVPath(), flag.Name))
	return nil
}

func getBenchmarkInst(benchFile string) (*m.Benchmark, error) {
	reqData, err := ioutil.ReadFile(param.GetBenchJsonPath(benchFile))
	if err != nil {
		log.Errorf(log.Benchmark, "Read bench conf file err:%v\n", err)
		return nil, fmt.Errorf("Read bench conf file err: %v", err)
	}

	var bench map[string][]m.Benchmark
	if err = json.Unmarshal(reqData, &bench); err != nil {
		log.Errorf(log.Benchmark, "Unmarshal err:%v\n", err)
		return nil, fmt.Errorf("Unmarshal bench conf file err: %v", err)
	}

	benchmarks := bench["benchmark"]
	if len(benchmarks) == 0 {
		log.Errorf(log.Benchmark, "benchmark json is null")
		return nil, fmt.Errorf("benchmark json is null")
	}

	tuneIP := strings.Split(config.KeenTune.TargetIP, ",")
	if len(tuneIP) == 0 {
		log.Errorf(log.Benchmark, "tune ip from acopsd.conf is empty")
		return nil, fmt.Errorf("tune ip from acopsd.conf is empty")
	}

	inst:= benchmarks[0]
	inst.Cmd = strings.Replace(strings.Replace(inst.Cmd, "{remote_script_path}", inst.FilePath, 1), "{target_ip}", tuneIP[0], 1)
	inst.Host = fmt.Sprintf("%s:%s", config.KeenTune.BenchIP, config.KeenTune.BenchPort)
	return &inst, nil
}