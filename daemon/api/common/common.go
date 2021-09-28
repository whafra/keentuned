package common

import (
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	utilhttp "keentune/daemon/common/utils/http"
	m "keentune/daemon/modules"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	Cmd  string
}

// DumpFlag ...
type DumpFlag struct {
	Name   string
	Output string
	Force  bool
}

var SystemRun bool

func IsDataNameUsed(name string) bool {
	dataList, _, err := GetDataList()
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

func GetDataList() ([]string, string, error) {
	resp, err := utilhttp.RemoteCall("GET", config.KeenTune.BrainIP+":"+config.KeenTune.BrainPort+"/sensitize_list", nil)
	if err != nil {
		return nil, "", err
	}

	var sensiList struct {
		Success bool       `json:"suc"`
		Data    []listInfo `json:"data"`
	}

	if err = json.Unmarshal(resp, &sensiList); err != nil {
		return nil, "", err
	}

	if !sensiList.Success {
		return nil, "", fmt.Errorf("remotecall sensitize_list return suc is false")
	}

	var paramString string
	var dataList []string

	for _, value := range sensiList.Data {
		paramString += fmt.Sprintf("%s,%s,%s;", value.Name, value.Scenario, value.Algorithm)
		dataList = append(dataList, value.Name)
	}

	return dataList, paramString, nil
}

func KeenTunedService(quit chan os.Signal) {
	// register router
	registerRouter()

	go func() {
		select {
		case <-quit:
			log.Info("", "keentune is interrupted")
			utilhttp.RemoteCall("GET", config.KeenTune.BrainIP+":"+config.KeenTune.BrainPort+"/end", nil)
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
		log.Errorf("","%v is not supported", flag.Cmd)
		return nil
	}

	inst := new(deleter)
	inst.cmd = flag.Cmd
	inst.fileName = fullName

	if err := inst.check(flag.Name); err != nil {
		log.Errorf(fmt.Sprintf("%s delete", inst.cmd), err.Error())
		return nil
	}
	
	if err := inst.delete(); err != nil {
		log.Errorf(fmt.Sprintf("%s delete", inst.cmd), err.Error())
		return nil
	}

	log.Infof(fmt.Sprintf("%s delete", inst.cmd), "[ok] %v delete successfully.", flag.Name)
	return nil
}

func (d *deleter) check(inputName string) error {
	if strings.Contains(d.fileName, strings.TrimRight(config.KeenTune.Home, "/")) {
		return fmt.Errorf("%v is not supported to delete", d.fileName)
	}

	if d.fileName == inputName {
		return fmt.Errorf("%v is non-existent", d.fileName)
	}

	if d.cmd == "param" {
		return nil
	}

	activeFileName := m.GetProfileWorkPath("active.conf")
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

func RollbackImpl(flag RollbackFlag, reply *string) error {
	url := config.KeenTune.TargetIP + ":" + config.KeenTune.TargetPort + "/rollback"
	err := utilhttp.ResponseSuccess("POST", url, nil)
	if err != nil {
		log.Errorf(fmt.Sprintf("%s rollback", flag.Cmd), "exec param rollback err: %v", err)
		*reply = log.ClientLogMap[fmt.Sprintf("%s rollback", flag.Cmd)]
		return err
	}

	return nil
}


func GetProfilePath(fileName string) string {
	workPath := m.GetProfileWorkPath(fileName)
	if file.IsPathExist(workPath) {
		return workPath
	}

	homePath := m.GetProfileHomePath() + fileName
	if file.IsPathExist(homePath) {
		return homePath
	}

	return fileName
}

func GetParameterPath(fileName string) string {
	workPath := m.GetTuningWorkPath(fileName)
	if file.IsPathExist(workPath) {
		return workPath
	}

	homePath := m.GetParamHomePath() + fileName
	if file.IsPathExist(homePath) {
		return homePath
	}

	return fileName
}