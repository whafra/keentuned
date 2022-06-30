package sensitize

import (
	com "keentune/daemon/api/common"
	"keentune/daemon/common/log"
)

type Service struct {
}

// Delete run sensitize delete service
func (s *Service) Delete(flag com.DeleteFlag, reply *string) error {
	defer func() {
		*reply = log.ClientLogMap[log.SensitizeDel]
		log.ClearCliLog(log.SensitizeDel)
	}()

	return com.RunTrainDelete(flag, reply)
}

