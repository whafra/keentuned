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
		return fmt.Errorf("operation does not support, job %v is running", m.GetRunningTask())
	}

	defer func() {
		*reply = log.ClientLogMap[log.ParamRollback]
		log.ClearCliLog(log.ParamRollback)
	}()

	detail, err := m.Rollback(log.ParamRollback, "param")
	if err != nil {
		return fmt.Errorf("%v", detail)
	}

	if detail != "" {
		log.Warn(log.ParamRollback, detail)
		return nil
	}

	log.Infof(log.ParamRollback, fmt.Sprintf("[ok] %v rollback successfully", flag.Cmd))
	return nil
}

