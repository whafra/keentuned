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

	err := m.Rollback(log.ParamRollback, "param")
	if err != nil {
		return fmt.Errorf("Rollback details:\n%v", err)
	}

	log.Infof(log.ParamRollback, fmt.Sprintf("[ok] %v rollback successfully", flag.Cmd))
	return nil
}
