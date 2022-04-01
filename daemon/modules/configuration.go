package modules

import (
	"encoding/json"
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"strings"
)

// Configuration define a group of parameter and benchmark score in this configuration.
type Configuration struct {
	Parameters []Parameter           `json:"parameters"`
	Score      map[string]ItemDetail `json:"score"`
	Round      int                   `json:"current_round"`
	budget     float32
	timeSpend  utils.TimeSpend
	targetIP   []string
}

type ReceivedConfigure struct {
	Candidate []Parameter           `json:"candidate"`
	Score     map[string]ItemDetail `json:"bench_score,omitempty"`
	Iteration int                   `json:"iteration"`
	Budget    float32               `json:"budget"`
}

type ItemDetail struct {
	Value    float32   `json:"value,omitempty"`
	Negative bool      `json:"negative"`
	Weight   float32   `json:"weight"`
	Strict   bool      `json:"strict"`
	Baseline []float32 `json:"base,omitempty"`
}

// Save configuration to profile file
func (conf Configuration) Save(fileName, suffix string) error {
	// acquire API return round is 1 less than the actual round value
	conf.Round += 1

	err := file.Dump2File(config.GetTuningWorkPath(fileName), fileName+suffix, conf)
	if err != nil {
		return err
	}
	return err
}

// collectParam collect param change map to struct map and state param success information
func collectParam(applyResp config.DBLMap) (string, map[string]Parameter, error) {
	var paramCollection = make(map[string]Parameter)
	var sucCount, failedCount int
	var failedInfoSlice [][]string

	if len(applyResp) == 0 {
		return "", nil, fmt.Errorf("apply response is null")
	}

	for domain, paramMap := range applyResp {
		for name, orgValue := range paramMap {
			var appliedInfo Parameter
			err := utils.Map2Struct(orgValue, &appliedInfo)
			if err != nil {
				return "", paramCollection, fmt.Errorf("collect Param:[%v]\n", err)
			}

			appliedInfo.DomainName = domain
			value, ok := appliedInfo.Value.(string)
			if ok && strings.Contains(value, "\t") {
				appliedInfo.Value = strings.ReplaceAll(value, "\t", " ")
			}

			paramCollection[name] = appliedInfo

			if appliedInfo.Success {
				sucCount++
				continue
			}

			failedCount++
			if failedCount == 1 {
				failedInfoSlice = append(failedInfoSlice, []string{"param name", "failed reason"})
			}
			failedInfoSlice = append(failedInfoSlice, []string{name, appliedInfo.Msg})
		}
	}

	var setResult string

	if failedCount == 0 {
		setResult = fmt.Sprintf("successed %v/%v", sucCount, sucCount)
		return setResult, paramCollection, nil
	}

	failedDetail := utils.FormatInTable(failedInfoSlice)
	setResult = fmt.Sprintf("successed %v/%v, failed %v; the failed details:%s", sucCount, sucCount+failedCount, failedCount, failedDetail)

	if failedCount == len(paramCollection) {
		return setResult, paramCollection, fmt.Errorf("return all failed: %v", failedDetail)
	}

	return setResult, paramCollection, nil
}

func getApplyResult(sucBytes []byte, id int) (config.DBLMap, error) {
	var applyShortRet struct {
		Success bool        `json:"suc"`
		Msg     interface{} `json:"msg"`
	}

	err := json.Unmarshal(sucBytes, &applyShortRet)
	if err != nil {
		return nil, err
	}

	if !applyShortRet.Success {
		return nil, fmt.Errorf("apply short return failed, msg:%v", applyShortRet.Msg)
	}

	var applyResp struct {
		Success bool          `json:"suc"`
		Data    config.DBLMap `json:"data"`
		Msg     interface{}   `json:"msg"`
	}

	select {
	case body := <-config.ApplyResultChan[id]:
		log.Debugf(log.ParamTune, "target id: %v receive apply result :[%v]\n", id, string(body))
		if err := json.Unmarshal(body, &applyResp); err != nil {
			return nil, fmt.Errorf("Parse apply response Unmarshal err: %v", err)
		}
	}

	if !applyResp.Success {
		return nil, fmt.Errorf("get apply result failed, msg: %v", applyResp.Msg)
	}

	return applyResp.Data, nil
}

func GetApplyResult(body []byte, id int) (string, map[string]Parameter, error) {
	applyResp, err := getApplyResult(body, id)
	if err != nil {
		return "", nil, err
	}

	return collectParam(applyResp)
}
