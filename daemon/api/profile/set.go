package profile

import (
	"fmt"
	com "keentune/daemon/api/common"
	"keentune/daemon/common/log"
	m "keentune/daemon/modules"
)

type SetFlag struct {
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
		return fmt.Errorf("operation does not support, job %v is running", m.GetRunningTask())
	}

	m.SetRunningTask(com.JobProfile, "set")
	defer func() {
		*reply = log.ClientLogMap[log.ProfSet]
		log.ClearCliLog(log.ProfSet)
		m.ClearTask()
	}()

	return SettingImpl(flag)
}

func SettingImpl(flag SetFlag) error {
	tuner := &m.Tuner{}

	tuner.Setter.Group = make([]bool, len(flag.Group))
	tuner.Setter.IdMap = make(map[int]int)
	tuner.Setter.ConfFile = make([]string, len(flag.ConfFile))
	copy(tuner.Setter.Group, flag.Group)
	copy(tuner.Setter.ConfFile, flag.ConfFile)

	return tuner.Set()
}

