package sensitize

import (
	"fmt"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	m "keentune/daemon/modules"
)

// Stop run sensitize stop service
func (s *Service) Stop(request string, reply *string) error {
	filePath := "/var/keentune/sensitize_jobs.csv"
	trainJob := file.GetRecord(filePath, "status", "running", "name")

	if trainJob != "" {
		file.UpdateRow(filePath, trainJob, map[int]interface{}{m.TrainStatusIdx: m.Stop})
		log.Warnf("", "Abort sensibility identification job '%v'.", trainJob)
		*reply = fmt.Sprintf("%v Abort sensibility identification job '%v'.\n", utils.ColorString("yellow", "[Warning]"), trainJob)
		//m.StopSig <- os.Interrupt
	} else {
		log.Infof("", "No training job needs to stop.")
		*reply = utils.ColorString("red", fmt.Sprintln("No training job needs to stop."))
	}

	return nil
}

