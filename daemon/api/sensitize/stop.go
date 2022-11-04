package sensitize

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

// Stop run sensitize stop service
func (s *Service) Stop(request string, reply *string) error {
	filePath := config.GetDumpPath(config.SensitizeCsv)
	trainJob := file.GetRecord(filePath, "status", "running", "name")

	if trainJob != "" {
		file.UpdateRow(filePath, trainJob, map[int]interface{}{m.TrainStatusIdx: m.Stop})

		stop()
		log.Warnf("", "Abort sensibility identification job '%v'.", trainJob)
		*reply = fmt.Sprintf("%v Abort sensibility identification job '%v'.\n", utils.ColorString("yellow", "[Warning]"), trainJob)
	} else {
		log.Infof("", "No training job needs to stop.")
		*reply = utils.ColorString("red", fmt.Sprintln("No training job needs to stop."))
	}

	return nil
}

func stop() {
	job := m.GetRunningTask()
	if job == "" {
		return
	}

	if strings.Split(job, " ")[0] == m.JobTraining {
		m.ClearTask()
		m.StopSig <- os.Interrupt
	}
}

