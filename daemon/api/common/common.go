package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	utilhttp "keentune/daemon/common/utils/http"
	"net/http"
	"os"
	"strings"
)

type listInfo struct {
	Name      string `json:"name"`
	Scenario  string `json:"type"` // enum:"collect", "tuning"
	Algorithm string `json:"algorithm"`
}

type deleter struct {
	cmd      string
	fileName string
}

type DeleteFlag struct {
	Name  string
	Cmd   string
	Force bool
}

type RollbackFlag struct {
	Cmd string
}

// DumpFlag ...
type DumpFlag struct {
	Name   string
	Output []string
	Force  bool
}

var activeJob = ""

var (
	JobTuning     = "tuning"
	JobCollection = "collect"
	JobTraining   = "train"
	JobBenchmark  = "benchmark"
)

func IsDataNameUsed(name string) bool {
	dataList, _, _, err := GetDataList()
	if err != nil {
		return false
	}

	for _, has := range dataList {
		if has == name {
			return true
		}
	}

	return false
}

func GetDataList() ([]string, string, string, error) {
	resp, err := utilhttp.RemoteCall("GET", config.KeenTune.BrainIP+":"+config.KeenTune.BrainPort+"/sensitize_list", nil)
	if err != nil {
		return nil, "", "", err
	}

	var sensiList struct {
		Success bool       `json:"suc"`
		Data    []listInfo `json:"data"`
	}

	if err = json.Unmarshal(resp, &sensiList); err != nil {
		return nil, "", "", err
	}

	if !sensiList.Success {
		return nil, "", "", fmt.Errorf("remotecall sensitize_list return suc is false")
	}

	var dataDetailSlice [][]string
	var dataList []string
	var collectList string

	if len(sensiList.Data) > 0 {
		dataDetailSlice = append(dataDetailSlice, []string{"data name", "application scenario", "algorithm"})
	}

	for _, value := range sensiList.Data {
		dataDetailSlice = append(dataDetailSlice, []string{value.Name, value.Scenario, value.Algorithm})
		dataList = append(dataList, value.Name)
		if value.Scenario == "collect" {
			collectList += fmt.Sprintf("\n\t%v", value.Name)
		}
	}

	return dataList, collectList, utils.FormatInTable(dataDetailSlice), nil
}

func KeenTunedService(quit chan os.Signal) {
	// register router
	registerRouter()

	go func() {
		select {
		case <-quit:
			log.Info("", "keentune is interrupted")
			if GetRunningTask() != "" {
				utilhttp.RemoteCall("GET", config.KeenTune.BrainIP+":"+config.KeenTune.BrainPort+"/end", nil)
			}
			os.Exit(1)
		}
	}()

	err := http.ListenAndServe(":"+config.KeenTune.Port, nil)
	if err != nil {
		log.Errorf("", "listen and serve err: %v", err)
		os.Exit(1)
	}
}

// RunDelete run delete file service
func RunDelete(flag DeleteFlag, reply *string) error {
	defer func() {
		*reply = log.ClientLogMap[fmt.Sprintf("%s delete", flag.Cmd)]
		log.ClearCliLog(fmt.Sprintf("%s delete", flag.Cmd))
	}()

	var fullName string
	switch flag.Cmd {
	case "param":
		fullName = GetParameterPath(flag.Name)
	case "profile":
		fullName = GetProfilePath(flag.Name)
	default:
		log.Errorf("", "%v is not supported", flag.Cmd)
		return nil
	}

	inst := new(deleter)
	inst.cmd = flag.Cmd
	inst.fileName = fullName

	if err := inst.check(flag.Name); err != nil {
		log.Errorf(fmt.Sprintf("%s delete", inst.cmd), err.Error())
		return fmt.Errorf("Check name failed: %v", err.Error())
	}

	if err := inst.delete(); err != nil {
		log.Errorf(fmt.Sprintf("%s delete", inst.cmd), err.Error())
		return fmt.Errorf("Delete failed: %v", err.Error())
	}

	log.Infof(fmt.Sprintf("%s delete", inst.cmd), "[ok] %v delete successfully", flag.Name)
	return nil
}

func (d *deleter) check(inputName string) error {
	if strings.Contains(d.fileName, config.KeenTune.Home) {
		return fmt.Errorf("%v is not supported to delete", d.fileName)
	}

	if d.cmd == "param" {
		return nil
	}

	activeFileName := config.GetProfileWorkPath("active.conf")
	if !file.IsPathExist(activeFileName) {
		return nil
	}

	activeNameBytes, err := ioutil.ReadFile(activeFileName)
	if err != nil {
		return fmt.Errorf("read file :%v err:%v", activeFileName, err)
	}

	if strings.Contains(d.fileName, string(activeNameBytes)) && string(activeNameBytes) != "" {
		return fmt.Errorf("%v is active profile, please run \"keentune profile rollback\" before delete", string(activeNameBytes))
	}

	return nil
}

func (d *deleter) delete() error {
	return os.RemoveAll(d.fileName)
}

func GetProfilePath(fileName string) string {
	if file.IsPathExist(fileName) {
		return fileName
	}

	workPath := config.GetProfileWorkPath(fileName)
	if file.IsPathExist(workPath) {
		return workPath
	}

	homePath := config.GetProfileHomePath(fileName)
	if file.IsPathExist(homePath) {
		return homePath
	}

	return ""
}

func GetParameterPath(fileName string) string {
	workPath := config.GetTuningWorkPath(fileName)
	if file.IsPathExist(workPath) {
		return workPath
	}

	homePath := config.GetParamHomePath() + fileName
	if file.IsPathExist(homePath) {
		return homePath
	}

	generatePath := config.GetGenerateWorkPath(fmt.Sprintf("%s%s", strings.TrimSuffix(fileName, ".json"), ".json"))
	if file.IsPathExist(generatePath) {
		return generatePath
	}

	return ""
}

func GetRunningTask() string {
	return activeJob
}

func SetRunningTask(class, name string) {
	activeJob = fmt.Sprintf("%s %s", class, name)
}

func ClearTask() {
	activeJob = ""
}

func IsJobRunning(name string) bool {
	return GetRunningTask() == name
}

func IsApplying() bool {
	job := GetRunningTask()
	if job == "" || len(strings.Split(job, " ")) < 2 {
		return false
	}

	return (strings.Split(job, " ")[0] == JobCollection) || (strings.Split(job, " ")[0] == JobTuning)
}

