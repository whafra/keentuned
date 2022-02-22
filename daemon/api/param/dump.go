package param

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	com "keentune/daemon/api/common"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	m "keentune/daemon/modules"
	"os"
	"reflect"
	"strings"
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

	tunedTaskPath := m.GetTuningWorkPath(dump.Name)
	outputFile := m.GetProfileWorkPath(dump.Output)
	err := checkDumpParam(tunedTaskPath, outputFile, dump.Force)
	if err != nil {
		log.Errorf(log.ParamDump, "Check dump param failed, err:%v", err)
		return fmt.Errorf("Check dump param failed, err:%v", err)
	}

	jsonFile := fmt.Sprintf("%s/%s_best.json", tunedTaskPath, dump.Name)
	if err = convertJsonFile2ConfigFile(jsonFile, outputFile); err != nil {
		log.Errorf(log.ParamDump, "Dump file failed, err:%v", err)
		return fmt.Errorf("Dump file failed, err:%v", err)
	}

	log.Infof(log.ParamDump, "[ok] %v dump successfully", outputFile)
	return nil
}

func checkDumpParam(path, outputFile string, confirm bool) error {
	if !file.IsPathExist(path) {
		return fmt.Errorf("find the tuned file [%v] does not exist, please confirm that the tuning job [%v] exists or is completed. ", path, strings.Split(path, "/")[len(strings.Split(path, "/"))-1])
	}

	activeFileName := m.GetProfileWorkPath("active.conf")
	if !file.IsPathExist(activeFileName) {
		fp, err := os.Create(activeFileName)
		if err != nil {
			return fmt.Errorf("create active file err:[%v]", err)
		}
		fp.Close()
	}

	if file.IsPathExist(outputFile) && !confirm {
		return fmt.Errorf("outputFile exist and you have given up to overwrite it")
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

	var proffileInfo string

	var paramsMap = make(map[string][]m.Parameter)

	for _, param := range configInfo.Parameters {
		paramsMap[param.DomainName] = append(paramsMap[param.DomainName], param)
	}

	index := 0
	for domain, params := range paramsMap {
		index++
		proffileInfo += fmt.Sprintf("[%s]\n", domain)

		for _, info := range params {
			if info.Dtype == "string" {
				proffileInfo += fmt.Sprintf("%s: \"%v\"\n", info.ParaName, info.Value)
				continue
			}

			num, ok := info.Value.(float64)
			if !ok {
				log.Warnf(log.ParamDump, "[%v] type %v is not float64", info.ParaName, reflect.TypeOf(info.Value))
				continue
			}
			proffileInfo += fmt.Sprintf("%s: %v\n", info.ParaName, uint(num))
		}

		if len(paramsMap) != index {
			proffileInfo += fmt.Sprintln()
		}
	}

	if err := ioutil.WriteFile(outputFile, []byte(proffileInfo), os.ModePerm); err != nil {
		return fmt.Errorf("write to file [%v] err: %v", outputFile, err)
	}

	return nil
}
