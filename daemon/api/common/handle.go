package common

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"net/http"
	"strings"
)

const (
	// LimitBytes
	LimitBytes = 1024 * 1024 * 5
)

func registerRouter() {
	http.HandleFunc("/benchmark_result", handler)
	http.HandleFunc("/apply_result", handler)
	http.HandleFunc("/sensitize_result", handler)
	http.HandleFunc("/status", status)
}

func handler(w http.ResponseWriter, r *http.Request) {
	// check request method
	var msg string
	if strings.ToUpper(r.Method) != "POST" {
		msg = fmt.Sprintf("request method [%v] is not found", r.Method)
		log.Error("", msg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(msg))
		return
	}

	bytes, err := ioutil.ReadAll(&io.LimitedReader{R: r.Body, N: LimitBytes})
	defer report(r.URL.Path, bytes, err)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"suc": true, "msg": ""}`))
	return
}

func report(url string, value []byte, err error) {
	if err != nil {
		msg := fmt.Sprintf("read request info err:%v", err)
		log.Error("", "report value to chan err:%v", msg)
	}

	if strings.Contains(url, "benchmark_result") && config.IsInnerBenchRequests[1] {
		var benchResult struct {
			BenchID int `json:"bench_id"`
		}
		err := json.Unmarshal(value, &benchResult)
		if err != nil {
			fmt.Printf("unmarshal bench id err: %v", err)
			return
		}

		config.BenchmarkResultChan[benchResult.BenchID] <- value

		return
	}

	if strings.Contains(url, "apply_result") {
		var applyResult struct {
			ID int `json:"target_id"`
		}
		err := json.Unmarshal(value, &applyResult)
		if err != nil {
			fmt.Printf("unmarshal apply target id err: %v", err)
			return
		}

		if config.IsInnerApplyRequests[applyResult.ID] && applyResult.ID > 0 {
			config.ApplyResultChan[applyResult.ID] <- value
		}

		return
	}

	if strings.Contains(url, "sensitize_result") && config.IsInnerSensitizeRequests[1] {
		config.SensitizeResultChan <- value
		return
	}
}

func status(w http.ResponseWriter, r *http.Request) {
	// check request method
	var msg string
	if strings.ToUpper(r.Method) != "GET" {
		msg = fmt.Sprintf("request method [%v] is not found", r.Method)
		log.Error("", msg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(msg))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "alive"}`))
	return
}
