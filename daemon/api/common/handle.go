package common

import (
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"fmt"
	"io"
	"io/ioutil"
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
	w.Write([]byte(`{"suc":true, "msg":"" }`))
	return
}

func report(url string, value []byte, err error) {
	if err != nil {
		msg := fmt.Sprintf("read request info err:%v", err)
		log.Error("","report value to chan err:%v",msg)
	}

	if strings.Contains(url, "benchmark_result") {
		config.BenchmarkResultChan <- value
	}

	if strings.Contains(url, "apply_result") {
		config.ApplyResultChan <- value
	}

	if strings.Contains(url, "sensitize_result") {
		config.SensitizeReusltChan <- value
	}
}