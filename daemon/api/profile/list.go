package profile

import (
	"fmt"
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
	records, _ := file.GetAllRecords(activeFileName)

	for _, value := range proFileList {
		if value == "active.conf" {
			continue
		}

		activeFlag := false
		for _, record := range records {
			if len(record) == 2 && record[0] == value {
				activeInfo := fmt.Sprintf("[active]\t%v", strings.Join(record, "\ttarget_info: "))
				activeFlag = true
				fileListInfo += fmt.Sprintln(utils.ColorString("GREEN", activeInfo))
				break
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

