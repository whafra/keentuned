package param

import (
	"fmt"
	com "keentune/daemon/api/common"
	"keentune/daemon/common/log"
	m "keentune/daemon/modules"
)

// Rollback run param rollback service
func (s *Service) Rollback(flag com.RollbackFlag, reply *string) error {
	if com.IsApplying() {
		return fmt.Errorf("operation does not support, job %v is running", com.GetRunningTask())
	}

	defer func() {
		*reply = log.ClientLogMap[log.ParamRollback]
		log.ClearCliLog(log.ParamRollback)
	}()

	result, err := m.Rollback(log.ParamRollback)
	if err != nil {
		return fmt.Errorf("%v", result)
	}

	if result != "" {
		log.Warn(log.ParamRollback, result)
		return nil
	}

	log.Infof(log.ParamRollback, fmt.Sprintf("[ok] %v rollback successfully", flag.Cmd))
	return nil
}

