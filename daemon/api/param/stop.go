package param

import (
	com "keentune/daemon/api/common"
	m "keentune/daemon/modules"
	"os"
	"strings"
	"fmt"
        "keentune/daemon/common/log"
)

// Stop run param stop service
func (s *Service) Stop(request string, reply *string) error {
	job := com.GetRunningTask()
	if job == "" {
                log.Errorf("", "No running job can stop.")
                return fmt.Errorf("No running job can stop.")
        }

	if job != "" && (strings.Split(job, " ")[0] == com.JobTuning || strings.Split(job, " ")[0] == com.JobBenchmark) {
		m.StopSig <- os.Interrupt
	}

	return nil
}
