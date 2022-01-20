package param

import (
	"fmt"
	com "keentune/daemon/api/common"
)

type Service struct {
}

// Delete run param delete service
func (s *Service) Delete(flag com.DeleteFlag, reply *string) error {
	if com.IsJobRunning(fmt.Sprintf("%s %s", com.JobTuning, flag.Name)) {
		return fmt.Errorf("tuning job %v is running, wait for it finishing", flag.Name)
	}

	return com.RunDelete(flag, reply)
}
