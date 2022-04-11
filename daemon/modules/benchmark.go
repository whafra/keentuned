package modules

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"keentune/daemon/common/utils/http"
	"time"
	"sync"
)

// Benchmark define benchmark cmd and host to run
type Benchmark struct {
	Cmd         string                `json:"benchmark_cmd"`
	Host        string                `json:"host"`
	FilePath    string                `json:"local_script_path"`
	Items       map[string]ItemDetail `json:"items"`
	round       int
	verbose     bool
	LogName     string   `json:"-"`
	SortedItems []string `json:"-"`
}

// BenchResult benchmark request result
type BenchResult struct {
	Success bool              `json:"suc"`
	Result  map[string]Result `json:"result,omitempty"`
	Message interface{}       `json:"msg,omitempty"`
}

type Result struct {
	Value float32 `json:"value"`
}

// RunBenchmark : run benchmark script or command in client
func (tuner *Tuner) RunBenchmark(num int) (map[string][]float32, map[string]ItemDetail, string, error) {	
	start := time.Now()
	var scores = make([]map[string][]float32, len(config.KeenTune.BenchGroup))
	var sumScore = make([]map[string]float32, len(config.KeenTune.BenchGroup))

	config.IsInnerBenchRequests[1] = true
	defer func() { config.IsInnerBenchRequests[1] = false }()

	var groupsScores = make([][]map[string]float32, num)
	for i := 1; i <= num; i++ {
		wg := sync.WaitGroup{}
		for groupID, group := range config.KeenTune.BenchGroup {
			groupsScores[groupID] = make([]map[string]float32, len(group.SrcIPs))
			scores[groupID] = make(map[string][]float32)
			sumScore[groupID] = make(map[string]float32)
			for index, benchIP := range group.SrcIPs {
				wg.Add(1)
				go func(wg *sync.WaitGroup, benchIP string, index int) {
					tuner.Benchmark.Host = fmt.Sprintf("%s:%s", benchIP, group.SrcPort)
					resp, err := http.RemoteCall("POST", tuner.Benchmark.Host+"/benchmark", getBenchReq(tuner.Benchmark.Cmd, benchIP))
					if err != nil {
						return
					}

					groupsScores[groupID][index], err = tuner.parseScore(resp, benchIP)
					if err != nil {
						return
					}
					wg.Done()

				}(&wg, benchIP, index)
			}
		}

		wg.Wait()

		//  collect score of each group
		for groupID, groupScores := range groupsScores {
			for _, results := range groupScores {
				for name, value := range results {
					scores[groupID][name] = append(scores[groupID][name], value)
					sumScore[groupID][name] += value
				}
			}
		}
	}

	tuner.Benchmark.round = num
	tuner.Benchmark.verbose = tuner.Verbose
	benchScoreResult, resultString, err := tuner.Benchmark.getScore(scores[0], sumScore[0], start, &tuner.timeSpend.benchmark)
	return scores[0], benchScoreResult, resultString, err
}

func getBenchReq(cmd, ip string) interface{} {
	var requestBody = map[string]interface{}{}
	requestBody["benchmark_cmd"] = cmd
	requestBody["resp_ip"] = config.RealLocalIP
	requestBody["resp_port"] = config.KeenTune.Port
	requestBody["bench_id"] = config.KeenTune.Bench.BenchIPMap[ip]
	return requestBody
}


func (benchmark Benchmark) getScore(scores map[string][]float32, sumScores map[string]float32, start time.Time, benchTime *time.Duration) (map[string]ItemDetail, string, error) {
	benchScoreResult := map[string]ItemDetail{}
	var average float32
	if len(scores) == 0 {
		return nil, "", fmt.Errorf("execute %v rounds all benchmark failed", benchmark.round)
	}

	if len(benchmark.Items) != len(scores) {
		log.Warnf("", "demand bench.json items length [%v] is not equal to benchmark api response scores length [%v], please check the bench.json and the python file you specified whether matched", len(benchmark.Items), len(scores))
	}

	resultString := ""
	for _, name := range benchmark.SortedItems {
		info, ok := benchmark.Items[name]
		if !ok {
			return nil, "", fmt.Errorf("assert item %v not found", name)
		}

		scoreSlice, ok := scores[name]
		if !ok {
			log.Warnf("", "benchmark response  [%v] detail info not exist, please check the bench.json and the python file you specified whether matched", name)
			continue
		}
		average = sumScores[name] / float32(len(scoreSlice))

		if benchmark.verbose {
			resultString += fmt.Sprintf("\n\t[%v]\t(weight: %.1f)\tscores %v,\taverage = %.3f,\t%v", name, info.Weight, scoreSlice, average, utils.Fluctuation(scoreSlice, average))
		}

		if !benchmark.verbose && info.Weight > 0.0 {
			resultString += fmt.Sprintf("\n\t[%v]\t(weight: %.1f)\taverage scores = %.3f", name, info.Weight, average)
		}

		benchScoreResult[name] = ItemDetail{
			Negative: info.Negative,
			Weight:   info.Weight,
			Strict:   info.Strict,
			Baseline: scoreSlice,
		}

	}

	timeCost := utils.Runtime(start)
	*benchTime += timeCost.Count

	if benchmark.verbose {
		resultString = fmt.Sprintf("%v, %v", resultString, timeCost.Desc)
	}

	return benchScoreResult, resultString, nil
}

// SendScript : send script file to client
func (benchmark Benchmark) SendScript(sendTime *time.Duration, Host string) (bool, string, error) {
	start := time.Now()
	benchBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", config.KeenTune.Home, benchmark.FilePath))
	if err != nil {
		return false, "", fmt.Errorf("SendScript readFile err:%v", err)
	}

	requestBody := map[string]interface{}{
		"file_name":   benchmark.FilePath,
		"body":        string(benchBytes),
		"encode_type": "utf-8",
	}

	err = http.ResponseSuccess("POST", Host+"/sendfile", requestBody)
	if err != nil {
		return false, "", fmt.Errorf("SendScript remote call err:%v", err)
	}

	timeCost := utils.Runtime(start)
	*sendTime += timeCost.Count

	return true, timeCost.Desc, nil
}

func (tuner *Tuner) parseScore(body []byte, ip string) (map[string]float32, error) {
	var benchResult BenchResult
	err := json.Unmarshal(body, &benchResult)
	if err != nil {
		return nil, fmt.Errorf("parse score err:%v", err)
	}

	if !benchResult.Success {
		return nil, fmt.Errorf("parse score failed, msg :%v", benchResult.Message)
	}

	var resultMap = map[string]float32{}
	id := config.KeenTune.Bench.BenchIPMap[ip]
	select {
	case bytes := <-config.BenchmarkResultChan[id]:
		log.Debugf("", "get benchmark result:%s", bytes)
		if err = json.Unmarshal(bytes, &benchResult); err != nil {
			return nil, fmt.Errorf("unmarshal request info err:%v", err)
		}

		if !benchResult.Success {
			return nil, fmt.Errorf("msg:%v", benchResult.Message)
		}

		for name, result := range benchResult.Result {
			resultMap[name] = result.Value
		}

		break
	case <-StopSig:
		tuner.rollback()
		return nil, fmt.Errorf("get benchmark is interrupted")
	}

	if len(resultMap) == 0 {
		return nil, fmt.Errorf("get benchmark result is nil")
	}

	return resultMap, nil
}

