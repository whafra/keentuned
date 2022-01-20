package param

import (
	"fmt"
	com "keentune/daemon/api/common"
	"keentune/daemon/common/log"
	m "keentune/daemon/modules"
)

// Rollback run param rollback service
func (s *Service) Rollback(flag com.RollbackFlag, reply *string) error {
	defer func() {
		*reply = log.ClientLogMap[log.ParamRollback]
		log.ClearCliLog(log.ParamRollback)
	}()

	result, allSuccess := m.Rollback(log.ParamRollback)
	if !allSuccess {
		return fmt.Errorf("Rollback details:\n%v", result)
	}

	log.Infof(log.ParamRollback, fmt.Sprintf("[ok] %v rollback successfully", flag.Cmd))
	return nil
}
