/*
Copyright © 2021 KeenTune

Package sensitize for daemon, this package contains the collect, delete, list, stop, train for sensitize parameter collection. Is a function implementation of the sensitize parameter collection server in the rpc framework.
*/
package sensitize

import (
	"fmt"
	"io/ioutil"
	com "keentune/daemon/api/common"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	m "keentune/daemon/modules"
	"os"
	"time"
)

type TrainFlag struct {
	Job    string
	Data   string
	Trials int
	Force  bool
	Log    string
	Config string
}

// Train run sensitize train service
func (s *Service) Train(flags TrainFlag, reply *string) error {
	err := config.Backup(flags.Config, flags.Job, "training")
	if err != nil {
		return fmt.Errorf("backup '%v' failed: %v", flags.Config, err)
	}

	if err := com.CheckBrainClient(); err != nil {
		return fmt.Errorf("check %v", err)
	}

	go runTrain(flags)
	return nil
}

func runTrain(flags TrainFlag) {
	m.SetRunningTask(com.JobTraining, flags.Job)
	log.SensitizeTrain = "sensitize train" + ":" + flags.Log
	ioutil.WriteFile(flags.Log, []byte{}, os.ModePerm)
	defer func() {
		m.ClearTask()
		config.ProgramNeedExit <- true
		<-config.ServeFinish
	}()

	log.Infof(log.SensitizeTrain, "Step1. Sensitize train data '%v' start, and algorithm is %v.", flags.Data, config.KeenTune.Explainer)

	if err := TrainImpl(flags, "training"); err != nil {
		log.Errorf(log.SensitizeTrain, "Param Tune failed, msg: %v", err)
		return
	}
	return
}

func TrainImpl(flag TrainFlag, cmd string) error {

	tuner := &m.Tuner{
		StartTime: time.Now(),
		Step:      1,
		Flag:      cmd,
		Algorithm: config.KeenTune.Explainer,
	}
	tuner.Trials = flag.Trials
	tuner.Data = flag.Data
	tuner.Job = flag.Job
	tuner.Train()
	return nil
}

