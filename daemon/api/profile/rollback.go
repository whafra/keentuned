package profile

import (
	com "keentune/daemon/api/common"
	"keentune/daemon/common/log"
	m "keentune/daemon/modules"
	"fmt"
)

// Rollback run profile rollback service
func (s *Service) Rollback(flag com.RollbackFlag, reply *string) error {
	defer func() {
		*reply = log.ClientLogMap[log.ProfRollback]
		log.ClearCliLog(log.ProfRollback)
	}()

	err := com.RollbackImpl(flag, reply)
	if err != nil {
		return err
	}

	fileName := m.GetProfileWorkPath("active.conf")
	if err := updateActiveFile(fileName, []byte{}); err != nil {
		log.Errorf(log.ProfRollback, "Update active file failed, err:%v", err)
		return fmt.Errorf("Update active file failed, err:%v", err)
	}

	log.Infof(log.ProfRollback, fmt.Sprintf("[ok] %v rollback successfully", flag.Cmd))
	return nil
}
