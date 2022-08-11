package system

import (
	"fmt"
	com "keentune/daemon/api/common"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	m "keentune/daemon/modules"
)

func (s *Service) RollbackAll(flag string, reply *string) error {
	result := new(string)
	if com.IsTargetOffline(result) {
		return fmt.Errorf("Find target: %v offline", *result)
	}

	if com.IsApplying() {
		return fmt.Errorf("operation does not support, job %v is running", m.GetRunningTask())
	}

	defer func() {
		*reply = log.ClientLogMap[log.RollbackAll]
		log.ClearCliLog(log.RollbackAll)
	}()

	detail, err := m.Rollback(log.RollbackAll, "original")
	if err != nil {
		return fmt.Errorf("%v", detail)
	}

	if detail != "" {
		log.Warn(log.RollbackAll, detail)
		return nil
	}

	log.Infof(log.RollbackAll, utils.ColorString("green",
		"[ok] Rollback all successfully"))
	return nil
}

