package sensitize

import (
	"fmt"
	"io/ioutil"
	com "keentune/daemon/api/common"
	"keentune/daemon/api/param"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"os"
)

type Service struct {
}

// Collect run sensitize collect service
func (s *Service) Collect(flag param.TuneFlag, reply *string) error {
	if com.GetRunningTask() != "" {
		log.Errorf("", "Job %v is running, you can wait for it finishing or stop it.", com.GetRunningTask())
		return fmt.Errorf("Job %v is running, you can wait for it finishing or stop it.", com.GetRunningTask())
	}

	if err := com.HeartbeatCheck(); err != nil {
		return fmt.Errorf("check %v", err)
	}

	go runCollect(flag)
	return nil
}

func runCollect(flag param.TuneFlag) {
	com.SetRunningTask(com.JobCollection, flag.Name)
	log.SensitizeCollect = "sensitize collect" + ":" + flag.Log
	ioutil.WriteFile(flag.Log, []byte{}, os.ModePerm)
	defer func() {
		config.ProgramNeedExit <- true
		<-config.ServeFinish
		com.ClearTask()
	}()

	log.Infof(log.SensitizeCollect, "Step1. Parameter auto Sensitize start, using algorithm = %v.", config.KeenTune.Sensitize.Algorithm)

	if err := param.TuningImpl(flag, "collect"); err != nil {
		log.Errorf(log.SensitizeCollect, "Sensitize Collect failed, msg:[%v]", err)
		return
	}
}
