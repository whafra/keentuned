package param

import (
	com "keentune/daemon/api/common"
	m "keentune/daemon/modules"
	"os"
	"strings"
)

// Stop run param stop service
func (s *Service) Stop(request string, reply *string) error {
	job := com.GetRunningTask()
	if job != "" && (strings.Split(job, " ")[0] == com.JobTuning || strings.Split(job, " ")[0] == com.JobBenchmark) {
		m.StopSig <- os.Interrupt
	}

	return nil
}
