package param

import (
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	m "keentune/daemon/modules"
	"os"
	"strings"
)

// Stop run param stop service
func (s *Service) Stop(request string, reply *string) error {
	filePath := config.GetDumpPath(config.TuneCsv)
	tuneJob := file.GetRecord(filePath, "status", "running", "name")

	if tuneJob != "" {
		file.UpdateRow(filePath, tuneJob, map[int]interface{}{m.TuneStatusIdx: m.Stop})
		stop()
		log.Warnf("", "Abort parameter optimization job '%v'.", tuneJob)
		*reply = fmt.Sprintf("%v Abort parameter optimization job '%v'.\n", utils.ColorString("yellow", "[Warning]"), tuneJob)
	} else {
		log.Infof("", "No tuning job needs to stop.")
		*reply = utils.ColorString("red", fmt.Sprintln("No tuning job needs to stop."))
	}

	return nil
}

func stop() {
	job := m.GetRunningTask()
	if job == "" {
		return
	}

	if strings.Split(job, " ")[0] == m.JobTuning || strings.Split(job, " ")[0] == m.JobBenchmark {
		m.ClearTask()
		m.StopSig <- os.Interrupt
	}
}

