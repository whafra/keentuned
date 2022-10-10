package main

import (
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	m "keentune/daemon/modules"
	"regexp"
	"strings"
)

func checkTrainingFlags(cmdName string, flag *TrainFlag) error {
	var err error
	jobFlag := "--job"

	if err = checkData(flag.Data); err != nil {
		return fmt.Errorf("'%v' %v", "--data", err)
	}

	if err = checkJob(cmdName, flag.Job); err != nil {
		return fmt.Errorf("'%v' %v", jobFlag, err)
	}

	if flag.Trials > 10 || flag.Trials < 1 {
		return fmt.Errorf("invalid value in trials=%v", flag.Trials)
	}

	if IsMutexJobRunning(m.JobTuning) {
		return fmt.Errorf("Another tuning job %v is running, use 'keentune param stop' to shutdown.", m.GetRunningTask())
	}

	return nil
}

// IsMutexJobRunning ...
func IsMutexJobRunning(mutexJob string) bool {
	job := m.GetRunningTask()
	if job == "" {
		return false
	}

	return strings.Split(job, " ")[0] == mutexJob
}

func checkData(name string) error {
	err := matchRegular(name)
	if err != nil {
		return err
	}

	var reason = new(string)
	if !isTuneDataReady(name, reason) {
		return fmt.Errorf("%v", *reason)
	}

	return nil
}

func isTuneDataReady(name string, reason *string) bool {
	filePath := config.GetDumpPath(config.TuneCsv)

	if !file.IsPathExist(filePath) {
		*reason = fmt.Sprintf("File '%v' does not exist.", filePath)
		return false
	}

	records, err := file.GetOneRecord(filePath, name, "name")
	if err != nil {
		*reason = fmt.Sprintf("The tuning job '%v' does not exist, run 'keentune param tune' to perform an auto-tuning job.", name)
		return false
	}

	if len(records) != m.TuneCols {
		*reason = fmt.Sprintf("invalid record '%v': column size %v, expected %v", name, len(records), m.TuneCols)
		return false
	}

	if records[m.TuneStatusIdx] == "finish" {
		return true
	}

	*reason = fmt.Sprintf("auto-tuning job '%v' is %v, can not training sensibility model", name, records[m.TuneStatusIdx])
	return false
}

func checkTuningFlags(cmdName string, flag *TuneFlag) error {
	var err error
	jobFlag := "--job"

	if err = checkJob(cmdName, flag.Name); err != nil {
		return fmt.Errorf("%v %v", jobFlag, err)
	}

	if flag.Round < 10 {
		return fmt.Errorf("invalid value: iteration = %v, Restriction: iteration >= 10 ", flag.Round)
	}

	if m.GetRunningTask() != "" {
		return fmt.Errorf("%v", parseJobRunningErrMsg(m.GetRunningTask()))
	}

	return nil
}

func parseJobRunningErrMsg(jobInfo string) string {
	parts := strings.Split(jobInfo, " ")
	if len(parts) != 2 {
		return fmt.Sprintf("Another job %v is running", jobInfo)
	}

	switch parts[0] {
	case m.JobTuning, m.JobBenchmark:
		return fmt.Sprintf("Another tuning job %v is running, use 'keentune param stop' to shutdown.", parts[1])
	case m.JobTraining:
		return fmt.Sprintf("Another sensitizing job %v is running, use 'keentune sensitize stop' to shutdown.", parts[1])
	default:
		return fmt.Sprintf("Another setting job %v is running, wait for it to finish.", parts[1])
	}
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
		command = m.JobTuning
		filePath = config.GetDumpPath(config.TuneCsv)
	}

	if cmd == "sensitize" {
		command = m.JobTraining
		filePath = config.GetDumpPath(config.SensitizeCsv)
	}

	if !file.IsPathExist(filePath) {
		*reason = fmt.Sprintf("File '%v' does not exist.", filePath)
		return false
	}

	if file.HasRecord(filePath, "name", name) {
		*reason = fmt.Sprintf("the specified name '%v' already exists, run '%v' first or change name and try again", name, getDeleteCmd(cmd))
		return true
	}

	// Mutual exclusion check
	runningJob := file.GetRecord(filePath, "status", "running", "name")
	if runningJob != "" {
		jobInfo := fmt.Sprintf("%s %s", command, runningJob)
		*reason = parseJobRunningErrMsg(jobInfo)
		return true
	}

	return false
}

func getDeleteCmd(cmd string) string {
	switch cmd {
	case "tune":
		return "keentune param delete"
	default:
		return fmt.Sprintf("keentune %v delete", cmd)
	}
}

