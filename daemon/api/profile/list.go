package profile

import (
	"fmt"
	"io/ioutil"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"strings"
)

// List run profile list service
func (s *Service) List(flag string, reply *string) error {
	defer func() {
		*reply = log.ClientLogMap[log.ProfList]
		log.ClearCliLog(log.ProfList)
	}()

	proFileList, err := walkProfileAllFiles()
	if err != nil {
		log.Errorf(log.ProfList, "Walk file path failed: %v", err)
		return fmt.Errorf("Walk file path failed: %v", err)
	}

	var fileListInfo string
	activeFileName := config.GetProfileWorkPath("active.conf")
	activeNameBytes, _ := ioutil.ReadFile(activeFileName)
	activeNames := strings.Split(string(activeNameBytes), "\n")

	for _, value := range proFileList {
		activeFlag := false
		for _, activeFile := range activeNames {
			if activeFile == value {
				activeFlag = true
				fileListInfo += fmt.Sprintf("%s\t%v\n", utils.ColorString("GREEN", "[active]"), value)
				continue
			}
			if value == "active.conf" {
				activeFlag = true
				continue
			}
		}
		if !activeFlag {
			fileListInfo += fmt.Sprintf("[available]\t%v\n", value)
		}
	}

	if len(fileListInfo) == 0 {
		log.Info(log.ProfList, "There is no profile configuration file exists, please execute keentune param dump first.")
		return nil
	}

	log.Infof(log.ProfList, "%v", strings.TrimSuffix(fileListInfo, "\n"))

	return nil
}

func walkProfileAllFiles() ([]string, error) {
	proFileList, err := file.WalkFilePath(config.GetProfileWorkPath(""), "", false)
	if err != nil {
		return proFileList, fmt.Errorf("walk dump folder failed :%v", err)
	}

	homeFileList, err := file.WalkFilePath(config.GetProfileHomePath(""), ".conf", false)
	if err != nil {
		return proFileList, fmt.Errorf("walk home folder failed :%v", err)
	}

	proFileList = append(proFileList, homeFileList...)
	return proFileList, nil
}

