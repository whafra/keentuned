package common

import (
	"encoding/json"
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	utilhttp "keentune/daemon/common/utils/http"
	m "keentune/daemon/modules"
	"net/http"
	"os"
	"strings"
	"syscall"
)

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

var (	
	tuningCsv    = "/var/keentune/tuning_jobs.csv"
	sensitizeCsv = "/var/keentune/sensitize_jobs.csv"
	logHome      = "/var/log/keentune"
)

var (
	JobTuning    = "tuning"
	JobProfile   = "profile"
	JobTraining  = "train"
	JobBenchmark = "benchmark"
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
		Success bool     `json:"suc"`
		Data    []string `json:"data"`
	}

	if err = json.Unmarshal(resp, &sensiList); err != nil {
		return nil, "", "", err
	}

	if !sensiList.Success {
		return nil, "", "", fmt.Errorf("remotecall sensitize_list return suc is false")
	}
	return sensiList.Data, "", "", nil
}

func KeenTunedService(quit chan os.Signal) {
	// register router
	registerRouter()

	go func() {
		select {
		case sig := <-quit:
			log.Info("", "keentune is interrupted")
			if m.GetRunningTask() != "" {
				killRunningJob()
				utilhttp.RemoteCall("GET", config.KeenTune.BrainIP+":"+config.KeenTune.BrainPort+"/end", nil)
			}
			if sig == syscall.SIGTERM {
				os.Exit(0)
			} else {
				os.Exit(1)
			}
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
		fullName = config.GetProfilePath(flag.Name)
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

	if flag.Cmd == "param" {
		primaryKeys := []string{flag.Name}
		file.DeleteRow(tuningCsv, primaryKeys)
		os.Remove(fmt.Sprintf("%v/%v.log", logHome, flag.Name))
	}

	log.Infof(fmt.Sprintf("%s delete", inst.cmd), "[ok] %v delete successfully", flag.Name)
	return nil
}

// RunDelete run delete file service
func RunTrainDelete(flag DeleteFlag, reply *string) error {
	fullName := GetSensitizePath(flag.Name)

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
	primaryKeys := []string{flag.Name}
	file.DeleteRow(sensitizeCsv, primaryKeys)

	log.Infof(log.SensitizeDel, "[ok] %v delete successfully", flag.Name)
	return nil
}

func (d *deleter) check(inputName string) error {
	if strings.Contains(d.fileName, config.KeenTune.Home) {
		return fmt.Errorf("%v is not supported to delete", d.fileName)
	}

	if d.cmd != "profile" {
		return nil
	}

	activeFileName := config.GetProfileWorkPath("active.conf")
	if !file.IsPathExist(activeFileName) {
		return nil
	}

	if file.HasRecord(activeFileName, "name", inputName) {
		return fmt.Errorf("%v is active profile, please run \"keentune profile rollback\" before delete", inputName)
	}

	return nil
}

func (d *deleter) delete() error {
	return os.RemoveAll(d.fileName)
}


func GetParameterPath(fileName string) string {
	workPath := config.GetTuningPath(fileName)
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

func GetSensitizePath(fileName string) string {
	workPath := config.GetSensitizePath(fileName)
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

func IsApplying() bool {
	job := m.GetRunningTask()
	if job == "" || len(strings.Split(job, " ")) < 2 {
		return false
	}

	return (strings.Split(job, " ")[0] == JobProfile) || (strings.Split(job, " ")[0] == JobTuning)
}

func killRunningJob() {
	m.ClearTask()

	tuningJob := file.GetRecord(tuningCsv, "status", "running", "name")
	if tuningJob != "" {
		file.UpdateRow(tuningCsv, tuningJob, map[int]interface{}{m.TuneStatusIdx: m.Kill})
	}

	sensitizeJob := file.GetRecord(sensitizeCsv, "status", "running", "name")
	if sensitizeJob != "" {
		file.UpdateRow(sensitizeCsv, sensitizeJob, map[int]interface{}{m.TrainStatusIdx: m.Kill})
	}
}

