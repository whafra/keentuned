package param

import (
	com "keentune/daemon/api/common"
	"keentune/daemon/common/log"
	"fmt"
)

// Rollback run param rollback service
func (s *Service) Rollback(flag com.RollbackFlag, reply *string) error {
	defer func() {
		*reply = log.ClientLogMap[log.ParamRollback]
		log.ClearCliLog(log.ParamRollback)
	}()

	err := com.RollbackImpl(flag, reply)
	if err != nil {
		return err
	}

	log.Infof(log.ParamRollback, fmt.Sprintf("[%v rollback] success", flag.Cmd))
	return nil
}
