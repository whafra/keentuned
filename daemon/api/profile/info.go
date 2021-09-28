package profile

import (
	"keentune/daemon/common/log"
	com "keentune/daemon/api/common"
	"io/ioutil"
)

// Info run profile info service
func (s *Service) Info(fileName string, reply *string) error {
	defer func() {
		*reply = log.ClientLogMap[log.ProfInfo]
		log.ClearCliLog(log.ProfInfo)
	}()

	fullName := com.GetProfilePath(fileName)
	if fullName == fileName {
		log.Errorf(log.ProfInfo, "%v is non-existent.", fileName)
		return nil
	}

	activeNameBytes, err := ioutil.ReadFile(fullName)
	if err != nil {
		log.Errorf(log.ProfInfo, "Read file :%v err:%v\n", fullName, err)
		return err
	}

	log.Infof(log.ProfInfo, "%s", activeNameBytes)
	return nil
}
