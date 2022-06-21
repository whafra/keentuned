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

	if err := com.HeartbeatCheck(); err != nil {
		return fmt.Errorf("check %v", err)
	}

	go runTrain(flags)
	return nil
}

func runTrain(flags TrainFlag) {
	log.SensitizeTrain = "sensitize train" + ":" + flags.Log
	ioutil.WriteFile(flags.Log, []byte{}, os.ModePerm)
	defer func() {
		config.ProgramNeedExit <- true
		<-config.ServeFinish
	}()

	log.Infof(log.SensitizeTrain, "Step1. Sensitize train data '%v' start, and algorithm is %v.", flags.Data, config.KeenTune.Sensitize.Algorithm)

	if err := TrainImpl(flags, "training"); err != nil {
		log.Errorf(log.ParamTune, "Param Tune failed, msg: %v", err)
		return
	}
	return
}

func TrainImpl(flag TrainFlag, cmd string) error {

	tuner := &m.Tuner{
		StartTime: time.Now(),
		Step:      1,
		Flag:      cmd,
		Algorithm: config.KeenTune.Sensitize.Algorithm,
	}
	tuner.Trials = flag.Trials
	tuner.Data = flag.Data
	tuner.Job = flag.Job
	tuner.Epoch = config.KeenTune.Sensitize.Epoch
	tuner.Train()
	return nil
}

