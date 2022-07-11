package main

import (
	"fmt"
	com "keentune/daemon/api/common"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	m "keentune/daemon/modules"
	"regexp"
	"strings"
)

func checkTrainingFlags(cmdName string, flag *TrainFlag) error {
	var err error
	jobFlag := "--job"

	if err = checkData(cmdName, flag.Data); err != nil {
		return fmt.Errorf("'%v' %v", "--data", err)
	}

	if err = checkJob(cmdName, flag.Job); err != nil {
		return fmt.Errorf("'%v' %v", jobFlag, err)
	}

	
	if flag.Trials > 10 || flag.Trials < 1 {
		return fmt.Errorf("invalid value in trials=%v", flag.Trials)
	}

	if IsMutexJobRunning(com.JobTuning) {
		return fmt.Errorf("Another Job %v is running, you can wait for it finishing or stop it.", m.GetRunningTask())
	}

	flag.Config = config.GetKeenTunedConfPath(flag.Config)
	if !file.IsPathExist(flag.Config) {
		fmt.Printf("config file '%v' does not exist", flag.Config)
	}

	return nil
}

func IsMutexJobRunning(mutexJob string) bool {
	job := m.GetRunningTask()
	if job == "" {
		return false
	}

	return strings.Split(job, " ")[0] == mutexJob
}

func checkData(cmd, name string) error {
	err := matchRegular(name)
	if err != nil {
		return err
	}

	if !com.IsDataNameUsed(name) {
		return fmt.Errorf("file '%v' does not exist", name)
	}

	var reason = new(string)
	if !isTuneDataReady(name, reason) {
		return fmt.Errorf("%v", *reason)
	}

	return nil
}

func isTuneDataReady(name string, reason *string) bool {
	command := ""
	filePath := ""
	command = "keentune param tune --job"
	filePath = config.GetDumpPath(config.TuneCsv)

	if !file.IsPathExist(filePath) {
		*reason = fmt.Sprintf("The file '%v' path does not exist", filePath)
		return false
	}

	records, err := file.GetOneRecord(filePath, name, "name")
	if err != nil {
		*reason = fmt.Sprintf("the specified name '%v' not exists. Run [%v %v --iteration xxx] or specify a new name and try again", name, command, name)
		return false
	}

	colLen := 11
	statusIdx := 2
	if len(records) != colLen {
		*reason = fmt.Sprintf("'%v' record is abnormal, column partial absence", name)
		return false
	}

	if records[statusIdx] == "finish" {
		return true
	}

	*reason = fmt.Sprintf("origin job '%v' status '%v' is not 'finish', training is not supported", name, records[statusIdx])
	return false
}

func checkTuningFlags(cmdName string, flag *TuneFlag) error {
	var err error
	jobFlag := "--job"

	if err = checkJob(cmdName, flag.Name); err != nil {
		return fmt.Errorf("%v %v", jobFlag, err)
	}

	if flag.Round < 10 {
		return fmt.Errorf("invalid value in iteration=%v, requirement: not less than 10", flag.Round)
	}

	if m.GetRunningTask() != "" {
		return fmt.Errorf("Job %v is running, you can wait for it finishing or stop it.", m.GetRunningTask())
	}

	flag.Config = config.GetKeenTunedConfPath(flag.Config)
	if !file.IsPathExist(flag.Config) {
		return fmt.Errorf("config file '%v' does not exist", flag.Config)
	}

	return nil
}

func checkJob(cmd, name string) error {
	err := matchRegular(name)
	if err != nil {
		return err
	}

	var reason = new(string)
	if isJobRepeatOrHasRunningJob(cmd, name, reason) {
		return fmt.Errorf("%v", *reason)
	}

	return nil
}

func matchRegular(name string) error {
	re := regexp.MustCompile("[^A-Za-z0-9_]+")
	re.FindAll([]byte(name), -1)

	var result string
	for _, value := range re.FindAllString(name, -1) {
		for _, v := range value {
			result += fmt.Sprintf("%q ", v)
		}
	}

	if len(result) != 0 {
		return fmt.Errorf("find unexpected characters %v. Only \"a-z\", \"A-Z\", \"0-9\" and \"_\" are supported", result)
	}
	return nil
}

func isJobRepeatOrHasRunningJob(cmd, name string, reason *string) bool {
	command := ""
	filePath := ""
	if cmd == "tune" {
		command = "keentune param delete --job"
		filePath = config.GetDumpPath(config.TuneCsv)
	}

	if cmd == "sensitize" {
		command = "keentune sensitize delete --job"
		filePath = config.GetDumpPath(config.SensitizeCsv)
	}

	if !file.IsPathExist(filePath) {
		*reason = fmt.Sprintf("The file '%v' path does not exist", filePath)
		return false
	}

	if file.HasRecord(filePath, "name", name) {
		*reason = fmt.Sprintf("the specified name '%v' already exists. Run [%v %v] or specify a new name and try again", name, command, name)
		return true
	}

	// Mutual exclusion check
	runningJob := file.GetRecord(filePath, "status", "running", "name")
	if runningJob != "" {
		*reason = fmt.Sprintf("Job %v is running, you can wait for finishing it or stop it.", runningJob)
		return true
	}

	return false
}

