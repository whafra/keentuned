package sensitize

import (
	com "keentune/daemon/api/common"
	"keentune/daemon/common/log"
	"fmt"
)

// List run sensitize list service
func (s *Service) List(flag string, reply *string) error {
	defer func() {
		*reply = log.ClientLogMap[log.SensitizeList]
		log.ClearCliLog(log.SensitizeList)
	}()

	_, _, sensiList, err := com.GetDataList()
	if err != nil {
		log.Errorf(log.SensitizeList, "Get sensitize Data List err:%v", err)
		return fmt.Errorf("Get sensitize Data List err:%v", err)
	}

	if len(sensiList) == 0 {
		log.Infof(log.SensitizeList, "No sensitive parameter identification record found, you can execute the command [keentune sensitize collect] first")
		return nil
	}

	log.Infof(log.SensitizeList, "Get sensitive parameter identification record successfully, and the details displayed in the terminal.")
	
	log.Infof(log.SensitizeList, "%s show table end.", fmt.Sprintf("%s;%s", "data name,application scenario,algorithm", sensiList))
	return nil
}
