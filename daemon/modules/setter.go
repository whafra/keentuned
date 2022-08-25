package modules

import (
	"fmt"
	"io/ioutil"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"os"
	"strings"
)

// Setter ...
type Setter struct {
	Group         []bool
	ConfFile      []string
	recommend     string
	preSetWarning string
}

// Set profile set  main process
func (tuner *Tuner) Set() {
	var err error
	tuner.logName = log.ProfSet
	if err = tuner.initProfiles(); err != nil {
		log.Errorf(log.ProfSet, "init profiles %v", err)
		return
	}

	if len(tuner.recommend) > 0 {
		fmtStr := fmt.Sprintf("%v\n%v\n", utils.ColorString("green", "[+] Recommendation (Manual Settings)"), tuner.recommend)
		log.Infof(log.ProfSet, fmtStr)
	}

	if len(tuner.preSetWarning) > 0 {
		log.Warn(log.ProfSet, tuner.preSetWarning)
	}

	defer func() {
		if err != nil {
			tuner.rollback()
		}
	}()

	if err = tuner.prepareBeforeSet(); err != nil {
		log.Error(log.ProfSet, err.Error())
		return
	}

	groupSetResult := fmt.Sprintf("%v\n", utils.ColorString("green", "[+] Profile Result (Auto Settings)"))
	if len(log.ClientLogMap[log.ProfSet]) > 0 {
		groupSetResult = fmt.Sprintf("\n%v", groupSetResult)
	}

	log.Infof(log.ProfSet, groupSetResult)

	err = tuner.setConfigure()
	if err != nil {
		log.Errorf(log.ProfSet, "Set failed: %v", err)
		return
	}

	err = tuner.updateActive()
	if err != nil {
		return
	}

	log.Info(log.ProfSet, tuner.applySummary)

	return
}

func (tuner *Tuner) updateActive() error {
	activeFile := config.GetProfileWorkPath("active.conf")
	// 先拼接，再写入
	var fileSet = fmt.Sprintln("name,group_info")
	var activeInfo = make(map[string][]string)
	for groupIndex, settable := range tuner.Setter.Group {
		if settable {
			fileName := file.GetPlainName(tuner.Setter.ConfFile[groupIndex])
			activeInfo[fileName] = append(activeInfo[fileName], fmt.Sprintf("group%v", groupIndex+1))
		}
	}

	for name, info := range activeInfo {
		fileSet += fmt.Sprintf("%s,%s\n", name, strings.Join(info, " "))
	}

	if err := UpdateActiveFile(activeFile, []byte(fileSet)); err != nil {
		log.Errorf(log.ProfSet, "Update active file err:%v", err)
		return fmt.Errorf("update active file err %v", err)
	}

	return nil
}

func (tuner *Tuner) prepareBeforeSet() error {
	// step1. rollback the target machine
	err := tuner.rollback()
	if err != nil {
		return fmt.Errorf("rollback failed:\n%v", tuner.rollbackFailure)
	}

	// step2. clear the active file
	fileName := config.GetProfileWorkPath("active.conf")
	if err = UpdateActiveFile(fileName, []byte{}); err != nil {
		return fmt.Errorf("update active file failed, err:%v", err)
	}

	// step3. backup the target machine
	err = tuner.backup()
	if tuner.backupWarning != "" {
		log.Warnf(tuner.logName, "%v", tuner.backupWarning)
	}

	if err != nil {
		return fmt.Errorf("%v", tuner.backupFailure)
	}
	return nil
}

// UpdateActiveFile ...
func UpdateActiveFile(fileName string, info []byte) error {
	if err := ioutil.WriteFile(fileName, info, os.ModePerm); err != nil {
		return err
	}

	return nil
}

