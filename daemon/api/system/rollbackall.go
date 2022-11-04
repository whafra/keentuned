package system

import (
	"fmt"
	com "keentune/daemon/api/common"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	m "keentune/daemon/modules"
	"strings"
)

func (s *Service) RollbackAll(flag string, reply *string) error {
	result := new(string)
	if com.IsTargetOffline(result) {
		return fmt.Errorf("Find target offline, details:\n%v", strings.TrimSuffix(*result, "\n"))
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
		return fmt.Errorf("Rollback All Failed:\n%v", strings.TrimSuffix(detail, "\n"))
	}

	if detail != "" {
		log.Warn(log.RollbackAll, detail)
		return nil
	}

	log.Infof(log.RollbackAll, utils.ColorString("green",
		"[ok] Rollback all successfully"))
	return nil
}

