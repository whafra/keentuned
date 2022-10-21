package param

import (
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
)

// List run param list service
func (s *Service) List(flag string, reply *string) error {
	defer func() {
		*reply = log.ClientLogMap[log.ParamList]
		log.ClearCliLog(log.ParamList)
	}()

	paramHeader := "Parameter List"
	benchHeader := "Benchmark List"

	paramInfo, err1 := walkAndShow(config.GetParamHomePath(), ".json", paramHeader)
	benchInfo, err2 := walkAndShow(config.GetBenchHomePath(), ".json", benchHeader)
	if err1 != nil || err2 != nil {
		log.Errorf(log.ParamList, "Walk path failed, param: %v and bench: %v", err1, err2)
		return nil
	}

	log.Infof(log.ParamList, "%v\n\n%v", paramInfo, benchInfo)

	return nil
}

func walkAndShow(filePath string, match string, header string) (string, error) {
	_, list, err := file.WalkFilePath(filePath, match)
	if err != nil {
		return "", fmt.Errorf("walk path: %v, err: %v", filePath, err)
	}

	return showList(list, header), nil
}

func showList(data []string, header string) string {
	var listInfo string
	for _, value := range data {
		listInfo += fmt.Sprintf("\n\t%v", value)
	}

	if listInfo != "" {
		return fmt.Sprintf("%v%v", header, listInfo)
	}

	return ""
}

