package param

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	com "keentune/daemon/api/common"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	m "keentune/daemon/modules"
	"os"
	"reflect"
)

// Dump run param dump service
func (s *Service) Dump(dump com.DumpFlag, reply *string) error {
	if com.IsJobRunning(fmt.Sprintf("%s %s", com.JobTuning, dump.Name)) {
		return fmt.Errorf("tuning job %v is running, wait for it finishing", dump.Name)
	}

	defer func() {
		*reply = log.ClientLogMap[log.ParamDump]
		log.ClearCliLog(log.ParamDump)
	}()

	tunedTaskPath := config.GetTuningWorkPath(dump.Name)
	var err error
	for index, outputFile := range dump.Output {
		jsonFile := fmt.Sprintf("%s/%s_group%v_best.json", tunedTaskPath, dump.Name, index+1)
		if err = convertJsonFile2ConfigFile(jsonFile, outputFile); err != nil {
			log.Errorf(log.ParamDump, "Dump file failed, err:%v", err)
			return fmt.Errorf("Dump file failed, err:%v", err)
		}
		log.Infof(log.ParamDump, "[ok] %v dump successfully", outputFile)
	}

	return nil
}

func convertJsonFile2ConfigFile(jsonFile, outputFile string) error {
	paraBytes, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		return fmt.Errorf("read file :%v err:%v", jsonFile, err)
	}

	configInfo := &m.Configuration{}
	if err = json.Unmarshal(paraBytes, configInfo); err != nil {
		return fmt.Errorf("Unmarshal err:%v", err)
	}

	var content string

	var paramsMap = make(map[string][]m.Parameter)

	for _, param := range configInfo.Parameters {
		paramsMap[param.DomainName] = append(paramsMap[param.DomainName], param)
	}

	index := 0
	for domain, params := range paramsMap {
		index++
		content += fmt.Sprintf("[%s]\n", domain)

		for _, info := range params {
			if info.Dtype == "string" {
				content += fmt.Sprintf("%s: \"%v\"\n", info.ParaName, info.Value)
				continue
			}

			num, ok := info.Value.(float64)
			if !ok {
				log.Warnf(log.ParamDump, "[%v] type %v is not float64", info.ParaName, reflect.TypeOf(info.Value))
				continue
			}
			content += fmt.Sprintf("%s: %v\n", info.ParaName, uint(num))
		}

		if len(paramsMap) != index {
			content += fmt.Sprintln()
		}
	}

	if err := ioutil.WriteFile(outputFile, []byte(content), os.ModePerm); err != nil {
		return fmt.Errorf("write to file [%v] err: %v", outputFile, err)
	}

	return nil
}
