package profile

import (
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	m "keentune/daemon/modules"
)

// GenFlag ...
type GenFlag struct {
	Name   string
	Output string
	Force  bool
}

// Generate run profile generate service
func (s *Service) Generate(flag GenFlag, reply *string) error {
	defer func() {
		*reply = log.ClientLogMap[log.ProfGenerate]
		log.ClearCliLog(log.ProfGenerate)
	}()

	fullName := config.GetProfileWorkPath(flag.Name)
	_, readMap, err := m.ConvertConfFileToJson(fullName)
	if err != nil {
		ResetValue(&fullName, config.GetProfileHomePath(flag.Name))
	}
	
	_, readMap, err = m.ConvertConfFileToJson(fullName)
	if err != nil {
		log.Errorf(log.ProfGenerate, "Convert file: %v, err:%v\n", flag.Name, err)
		return fmt.Errorf("Convert file: %v, err:%v", flag.Name, err)
	}

	m.UpdateParams(readMap)

	if err := file.Dump2File(config.GetGenerateWorkPath(""), flag.Output, readMap); err != nil {
		log.Errorf(log.ProfGenerate, "Dump config info to json file [%v] err: %v", flag.Output, err)
		return fmt.Errorf("Dump json file: %v, err: %v", flag.Output, err)
	}

	log.Infof(log.ProfGenerate, "[ok] %v generate successfully", config.GetGenerateWorkPath(flag.Output))
	return nil
}

//ResetValue modify string value
func ResetValue(s *string, newValue string) {
	sByte := []byte(*s)
	for i := 0; i < len(sByte); i++ {
		sByte[i] = ' '
	}
	*s = newValue
}

