package param

import (
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	m "keentune/daemon/modules"
	"fmt"
)

// List run param list service
func (s *Service) List(flag string, reply *string) error {
	defer func() {
		*reply = log.ClientLogMap[log.ParamList]
		log.ClearCliLog(log.ParamList)
	}()

	tuneList, err := walkParamAllPaths()
	if err != nil {
		log.Errorf(log.ParamList, "msg:%v", err)
		return fmt.Errorf("[Error] msg:%v", err)
	}

	if len(tuneList) == 0 {
		log.Info(log.ParamList, "There is no param file existence, please execute keentune param tune first.")
		return nil
	}

	var tuneInfo string

	for _, value := range tuneList {
		tuneInfo += fmt.Sprintf("\n\t%v", value)
	}

	log.Infof(log.ParamList, "Param list as follows:%v", tuneInfo)
	return nil
}

func walkParamAllPaths() ([]string, error) {
	tuneList, err := file.WalkFilePath(m.GetTuningWorkPath("") + "/", "", true)
	if err != nil {
		return tuneList, fmt.Errorf("walk tune work path:%v", err)
	}

	homeList, err := file.WalkFilePath(m.GetParamHomePath(), "", false)
	if err != nil {
		return tuneList, fmt.Errorf("walk param home path:%v", err)
	}

	tuneList = append(tuneList, homeList...)
	return tuneList, nil
}
