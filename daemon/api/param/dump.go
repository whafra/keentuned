package param

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	com "keentune/daemon/api/common"
	"keentune/daemon/common/log"
	m "keentune/daemon/modules"
	"reflect"
	"strings"
)

// Dump run param dump service
func (s *Service) Dump(dump com.DumpFlag, reply *string) error {
	defer func() {
		*reply = log.ClientLogMap[log.ParamDump]
		log.ClearCliLog(log.ParamDump)
	}()

	var err error
	var dumpFiles string
	for _, combination := range dump.Output {
		parts := strings.Split(combination, ",")
		if len(parts) != 2 {
			return fmt.Errorf("find combind name '%v' abnormal", combination)
		}

		if err = convertJsonFile2ConfigFile(parts[0], parts[1]); err != nil {
			log.Errorf(log.ParamDump, "Dump file failed, err:%v", err)
			return fmt.Errorf("Dump file failed, err:%v", err)
		}

		dumpFiles += fmt.Sprintf("\n\t%v", parts[1])
	}

	log.Infof(log.ParamDump, "[ok] dump successfully, file list:%v", dumpFiles)

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

	if err := ioutil.WriteFile(outputFile, []byte(content), 0666); err != nil {
		return fmt.Errorf("write to file [%v] err: %v", outputFile, err)
	}

	return nil
}
