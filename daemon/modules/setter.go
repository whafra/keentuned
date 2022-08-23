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
	Group     []bool
	ConfFile  []string
	recommend string
}

// Set profile set  main process
func (tuner *Tuner) Set() error {
	var err error
	tuner.logName = log.ProfSet
	if err = tuner.initProfiles(); err != nil {
		log.Errorf(log.ProfSet, "init profiles %v", err)
		return fmt.Errorf("init profiles %v", err)
	}

	if len(tuner.recommend) > 0 {
		fmtStr := fmt.Sprintf("%v\n\n%v", utils.ColorString("green", "[+] Optimization Recommendation (Manual Setting)"), tuner.recommend)
		log.Infof(log.ProfSet, fmtStr)
	}

	defer func() {
		if err != nil {
			tuner.rollback()
		}
	}()

	if err = tuner.prepareBeforeSet(); err != nil {
		log.Errorf(log.ProfSet, "prepare for setting %v", err)
		return fmt.Errorf("prepare for setting %v", err)
	}

	if tuner.backupWarning != "" {
		log.Warnf(log.ProfSet, "%v", tuner.backupWarning)
	}

	err = tuner.setConfigure()
	if err != nil {
		log.Errorf(log.ProfSet, "Set failed: %v", err)
		return err
	}

	err = tuner.updateActive()
	if err != nil {
		return err
	}

	groupSetResult := fmt.Sprintf("%v\n\n", utils.ColorString("green", "[+] Optimization Result (Auto Setting)"))
	groupSetResult += strings.TrimSuffix(tuner.applySummary, "\n")
	log.Infof(log.ProfSet, groupSetResult)

	return nil
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
	if err != nil {
		return fmt.Errorf("backup failed:\n%v", tuner.backupFailure)
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

