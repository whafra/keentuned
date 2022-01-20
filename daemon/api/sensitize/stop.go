package sensitize

import (
	com "keentune/daemon/api/common"
	m "keentune/daemon/modules"
	"os"
	"strings"
)

// Stop run sensitize stop service
func (s *Service) Stop(request string, reply *string) error {
	job := com.GetRunningTask()
	if job != "" && (strings.Split(job, " ")[0] == com.JobCollection || strings.Split(job, " ")[0] == com.JobTraining) {
		m.StopSig <- os.Interrupt
	}

	return nil
}
