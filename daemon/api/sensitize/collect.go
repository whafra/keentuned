package sensitize

import (
	com "keentune/daemon/api/common"
	"keentune/daemon/api/param"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
)

type Service struct {
}

// Collect run sensitize collect service
func (s *Service) Collect(flag param.TuneFlag, reply *string) error {
	go runCollect(flag)
	return nil
}
func runCollect(flag param.TuneFlag) {
	if com.SystemRun {
		log.Info(log.SensitizeCollect, "An instance is running, please wait for it to finish and re-initiate the request.")
		return
	}

	com.SystemRun = true
	log.ClearCliLog(log.SensitizeCollect)

	defer func() {
		config.ProgramNeedExit <- true
		<-config.ServeFinish
		com.SystemRun = false
	}()

	go com.HeartbeatCheck()
	log.Infof(log.SensitizeCollect, "\nStep1. Parameter auto Sensitize start, using algorithm = %v.\n", config.KeenTune.Sensitize.Algorithm)

	if com.IsDataNameUsed(flag.Name) {
		log.Errorf(log.SensitizeCollect, "Please check the data name specified, already exists")
		return
	}

	if err := param.TuningImpl(flag, "collect"); err != nil {
		return
	}
}
