package profile

import (
	"fmt"
	com "keentune/daemon/api/common"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	m "keentune/daemon/modules"
)

// Rollback run profile rollback service
func (s *Service) Rollback(flag com.RollbackFlag, reply *string) error {
	if com.IsApplying() {
		return fmt.Errorf("operation does not support, job %v is running", m.GetRunningTask())
	}

	defer func() {
		*reply = log.ClientLogMap[log.ProfRollback]
		log.ClearCliLog(log.ProfRollback)
	}()

	detail, err := m.Rollback(log.ProfRollback, "profile")
	if err != nil {
		return fmt.Errorf("%v", detail)
	}

	fileName := config.GetProfileWorkPath("active.conf")
	if err := m.UpdateActiveFile(fileName, []byte{}); err != nil {
		log.Errorf(log.ProfRollback, "Update active file failed, err:%v", err)
		return fmt.Errorf("Update active file failed, err:%v", err)
	}

	if detail != "" {
		log.Warn(log.ProfRollback, detail)
		return nil
	}

	log.Infof(log.ProfRollback, fmt.Sprintf("[ok] %v rollback successfully", flag.Cmd))
	return nil
}

