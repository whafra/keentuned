package modules

import (
	"encoding/json"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils/http"
)

type backupDetail struct {
	// Available ...
	Available bool        `json:"avaliable"`
	Value     interface{} `json:"value"`
	Msg       interface{} `json:"msg"`
}

func (tuner *Tuner) backup() error {
	err := tuner.concurrent("backup", true)
	if tuner.backupWarning != "" {
		log.Warnf(tuner.logName, "Remove backup failure parameters:\n%v", tuner.backupWarning)
	}

	if err != nil {
		return err
	}

	if tuner.Flag == JobProfile {
		return nil
	}

	if tuner.backupWarning != "" {
		tuner.deleteUnAVLParams()
	}

	return nil
}

func callBackup(method, url string, request interface{}) (map[string]map[string]string, string, int) {
	var response map[string]map[string]backupDetail
	resp, err := http.RemoteCall(method, url, request)
	if err != nil {
		return nil, err.Error(), FAILED
	}

	if err = json.Unmarshal(resp, &response); err != nil {
		return nil, err.Error(), FAILED
	}

	req, ok := request.(map[string]interface{})
	if !ok {
		return nil, "assert request type to map error", FAILED
	}

	var unAVLParam = make(map[string]map[string]string)

	for domain, param := range req {
		_, notExist := unAVLParam[domain]
		if !notExist {
			unAVLParam[domain] = make(map[string]string)
		}
		parameter := param.(map[string]interface{})
		for name, _ := range parameter {
			_, exists := response[domain][name]
			if !exists || !response[domain][name].Available {
				msgBytes, _ := json.Marshal(response[domain][name].Msg)
				msg := string(msgBytes)
				if msg == "" || msg == "null" {
					msg = "domain can not backup"
				}

				unAVLParam[domain][name] = msg
			}
		}
	}

	return unAVLParam, "", SUCCESS
}

