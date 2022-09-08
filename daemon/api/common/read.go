package common

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	m "keentune/daemon/modules"
	"net/http"
	"strconv"
	"strings"
)

// TuneCmdResp ...
type TuneCmdResp struct {
	Iteration    int    `json:"iteration"`
	BaseRound    int    `json:"baseline_bench_round"`
	TuningRound  int    `json:"tuning_bench_round"`
	RecheckRound int    `json:"recheck_bench_round"`
	Algo         string `json:"algorithm"`
}

// TrainCmdResp ...
type TrainCmdResp struct {
	Trial int    `json:"trial"`
	Algo  string `json:"algorithm"`
	Data  string `json:"data"`
}

func read(w http.ResponseWriter, r *http.Request) {
	var result = new(string)
	w.Header().Set("content-type", "text/json")
	if strings.ToUpper(r.Method) != "POST" {
		*result = fmt.Sprintf("request method '%v' is not supported", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(*result))
		return
	}

	var err error
	defer func() {
		w.WriteHeader(http.StatusOK)
		if err != nil {
			w.Write([]byte(fmt.Sprintf("{\"suc\": false, \"msg\": \"%v\"}", err.Error())))
			log.Errorf("", "read operation: %v", err)
			return
		}

		w.Write([]byte(fmt.Sprintf("{\"suc\": true, \"msg\": %s}", *result)))
	}()

	bytes, err := ioutil.ReadAll(&io.LimitedReader{R: r.Body, N: LimitBytes})
	if err != nil {
		return
	}

	var req struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}

	err = json.Unmarshal(bytes, &req)
	if err != nil {
		err = fmt.Errorf("parse request info failed: %v", err)
		return
	}

	switch strings.ToLower(req.Type) {
	case "tuning":
		err = readTuneInfo(req.Name, result)
		return
	case "training":
		err = readTrainInfo(req.Name, result)
		return
	case "param-bench":
		err = readConfigParam(req.Name, result)
		return
	default:
		err = fmt.Errorf("type '%v' is not supported", req.Type)
		return
	}
}

func readConfigParam(job string, result *string) error {
	params, bench, err := config.GetJobParamConfig(job)
	if err != nil {
		return err
	}
	var resp struct {
		Params string `json:"parameters"`
		Bench  string `json:"benchmark"`
	}

	resp.Params = params
	resp.Bench = bench

	bytes, err := json.Marshal(resp)
	* result = string(bytes)
	return err
}

func readTrainInfo(job string, result *string) error {
	record, err := file.GetOneRecord(config.GetDumpPath(config.SensitizeCsv), job, "name")
	if err != nil {
		return fmt.Errorf("search '%v' err: %v", job, err)
	}

	if len(record) != m.TrainCols {
		return fmt.Errorf("search '%v' record '%v' Incomplete", job, strings.Join(record, ","))
	}

	var resp TrainCmdResp
	resp.Trial, _ = strconv.Atoi(record[m.TrainTrialsIdx])
	resp.Algo = record[m.TrainAlgoIdx]
	resp.Data = record[m.TrainDataIdx]
	bytes, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	*result = string(bytes)
	return nil
}

func parseBenchRound(info string, resp *TuneCmdResp) error {
	if strings.Contains(strings.ToLower(info), "baseline_bench_round") {
		num, err := parseRound(info, "baseline_bench_round")
		if err != nil {
			return err
		}

		resp.BaseRound = num
	}

	if strings.Contains(strings.ToLower(info), "tuning_bench_round") {
		num, err := parseRound(info, "tuning_bench_round")
		if err != nil {
			return err
		}
		resp.TuningRound = num
	}

	if strings.Contains(strings.ToLower(info), "recheck_bench_round") {
		num, err := parseRound(info, "recheck_bench_round")
		if err != nil {
			return err
		}

		resp.RecheckRound = num
	}

	return nil
}

func parseRound(info, key string) (int, error) {
	if !strings.Contains(strings.ToLower(info), key) {
		return 0, nil
	}

	flagParts := strings.Split(info, "=")
	if len(flagParts) != 2 {
		return 0, fmt.Errorf("algorithm not found")
	}

	num, err := strconv.Atoi(strings.TrimSpace(flagParts[1]))
	if err != nil {
		return 0, fmt.Errorf("get %v number err %v", key, err)
	}

	return num, nil
}

func readTuneInfo(job string, result *string) error {
	records, err := file.GetOneRecord(config.GetDumpPath(config.TuneCsv), job, "name")
	if err != nil {
		return fmt.Errorf("'%v' not exists", job)
	}

	if len(records) != m.TuneCols {
		return fmt.Errorf("invalid record '%v': column size %v, expected %v", job, len(records), m.TuneCols)
	}

	var resp = TuneCmdResp{}
	iteration, err := strconv.Atoi(strings.Trim(records[m.TuneRoundIdx], " "))
	if err != nil || iteration <= 0 {
		return fmt.Errorf("'%v' not exists", "iteration")
	}

	resp.Algo = records[m.TuneAlgoIdx]
	resp.Iteration = iteration

	m, err := config.GetRerunConf(records[m.TuneWSPIdx] + "/keentuned.conf")
	if err != nil {
		return err
	}

	resp.BaseRound = m["BaseRound"].(int)
	resp.TuningRound = m["TuningRound"].(int)
	resp.RecheckRound = m["RecheckRound"].(int)

	bytes, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	*result = string(bytes)
	return nil
}

