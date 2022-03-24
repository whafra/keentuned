package sensitize

import (
	"fmt"
	com "keentune/daemon/api/common"
	"keentune/daemon/common/log"
	"strings"
)

// List run sensitize list service
func (s *Service) List(flag string, reply *string) error {
	defer func() {
		*reply = log.ClientLogMap[log.SensitizeList]
		log.ClearCliLog(log.SensitizeList)
	}()

	_, _, sensiList, err := com.GetDataList()
	if err != nil {
		if find := strings.Contains(err.Error(), "connection refused"); find {
                        return fmt.Errorf("brain access denied")
                }
		log.Errorf(log.SensitizeList, "Get sensitize Data List err:%v", err)
		return fmt.Errorf("Get sensitize Data List err:%v", err)
	}

	if len(sensiList) == 0 {
		log.Infof(log.SensitizeList, "No sensitive parameter identification record found, you can execute the command [keentune sensitize collect] first")
		return nil
	}

	log.Infof(log.SensitizeList, "Get sensitive parameter identification results successfully, and the details displayed in the terminal.%v", sensiList)

	return nil
}

