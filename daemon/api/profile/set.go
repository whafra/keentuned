/*
Copyright © 2021 KeenTune

Package profile for daemon, this package contains the delete, generate, info, list, rollback, set for static tuning. Is a function implementation of the static tuning server in the rpc framework.
*/
package profile

import (
	"fmt"
	com "keentune/daemon/api/common"
	"keentune/daemon/common/log"
	m "keentune/daemon/modules"
	"strings"
)

// SetFlag ...
type SetFlag struct {
	Group    []bool
	ConfFile []string
}

// Result ...
type Result struct {
	Info    string
	Success bool
}

// Set run profile set service
func (s *Service) Set(flag SetFlag, reply *string) error {
	if com.IsApplying() {
		return fmt.Errorf("operation does not support, job %v is running", m.GetRunningTask())
	}

	var targetMsg = new(string)
	if com.IsSetTargetOffline(flag.Group, targetMsg) {
		return fmt.Errorf("found %v offline, please get them (it) ready before setting", strings.TrimSuffix(*targetMsg, ", "))
	}

	m.SetRunningTask(m.JobProfile, "set")
	defer func() {
		*reply = log.ClientLogMap[log.ProfSet]
		log.ClearCliLog(log.ProfSet)
		m.ClearTask()
	}()

	SettingImpl(flag)
	return nil
}

// SettingImpl ...
func SettingImpl(flag SetFlag) {
	tuner := &m.Tuner{}

	tuner.Setter.Group = make([]bool, len(flag.Group))
	tuner.Setter.ConfFile = make([]string, len(flag.ConfFile))
	copy(tuner.Setter.Group, flag.Group)
	copy(tuner.Setter.ConfFile, flag.ConfFile)
	tuner.Flag = m.JobProfile
	tuner.Set()
}

