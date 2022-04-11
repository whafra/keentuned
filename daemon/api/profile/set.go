package profile

import (
	"fmt"
	com "keentune/daemon/api/common"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	m "keentune/daemon/modules"
	"time"
)

type SetFlag struct {
	Name     string
	Group    []bool
	ConfFile []string
}

type Result struct {
	Info    string
	Success bool
}

// Set run profile set service
func (s *Service) Set(flag SetFlag, reply *string) error {
	if com.IsApplying() {
		return fmt.Errorf("operation does not support, job %v is running", com.GetRunningTask())
	}
	if err := com.HeartbeatCheck(); err != nil {
		log.Errorf(log.ProfSet, "Check %v", err)
		return fmt.Errorf("Check %v", err)
	}
	runSeting(flag, reply)
	return nil
}

func runSeting(flag SetFlag, reply *string) {

	defer func() {
		config.ProgramNeedExit <- true
		<-config.ServeFinish
		*reply = log.ClientLogMap[log.ProfSet]
		log.ClearCliLog(log.ProfSet)
	}()
	//log.Infof(log.ProfSet, "Step1. Profile Set start\n")
	if err := SetingImpl(flag, "tuning"); err != nil {
		log.Errorf(log.ProfSet, "Profile Set failed, msg: %v", err)
		return
	}

}

func SetingImpl(flag SetFlag, cmd string) error {

	tuner := &m.Tuner{
		Name:      flag.Name,
		StartTime: time.Now(),
		Flag:      cmd,
		Step:      1,
	}
	tuner.Seter.Group = make([]bool, len(flag.Group))
	tuner.Seter.ConfFile = make([]string, len(flag.ConfFile))
	copy(tuner.Seter.Group, flag.Group)
	copy(tuner.Seter.ConfFile, flag.ConfFile)

	tuner.Set()
	return nil
}
