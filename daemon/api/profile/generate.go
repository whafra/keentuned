package profile

import (
	"keentune/daemon/common/config"
	com "keentune/daemon/api/common"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	m "keentune/daemon/modules"
)

// Generate run profile generate service
func (s *Service) Generate(flag com.DumpFlag, reply *string) error {
	defer func() {
		*reply = log.ClientLogMap[log.ProfGenerate]
		log.ClearCliLog(log.ProfGenerate)
	}()
	if file.IsPathExist(m.GetTuningWorkPath(flag.Output)) && !flag.Force {
		log.Errorf(log.ProfGenerate, "outputFile %v exist and you have given up to overwrite it\n", flag.Name)
		return nil
	}

	fullName := m.GetProfileWorkPath(flag.Name)
	readMap, err := file.ConvertConfFileToJson(fullName)
	if err != nil {
		log.Errorf(log.ProfGenerate, "convert file %v to json err:%v\n", flag.Name, err)
		return err
	}

	totalParamMap, err := file.ReadFile2Map(config.KeenTune.Home + config.ParamAllFile)
	if err != nil {
		log.Errorf(log.ProfGenerate, "read [%v] file err:%v\n", config.KeenTune.Home+config.ParamAllFile, err)
		return err
	}

	_, _ = m.AssembleParams(readMap, totalParamMap)

	if err := file.Dump2File(m.GetTuningWorkPath(flag.Output), flag.Output+".json", readMap); err != nil {
		log.Errorf(log.ProfGenerate, "dump config info to json file [%v] err: %v", flag.Output, err)
		return err
	}

	log.Infof(log.ProfGenerate, "Generate file %v successfully.", m.GetTuningWorkPath(flag.Output)+"/"+flag.Output+".json")
	return nil
}
