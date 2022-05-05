package main

import (
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"regexp"
)

func checkTrainingFlags(cmdName string, flag *TrainFlag) error {
	var err error
	jobFlag := "--data"
	if cmdName == "train" {
		jobFlag = "--job"
	}

	if err = checkData(cmdName, flag.Data); err != nil {
		return fmt.Errorf("%v %v", flag.Data, err)
	}

	if err = checkJob(cmdName, flag.Job); err != nil {
		return fmt.Errorf("%v %v", jobFlag, err)
	}

	if flag.Trials <= 0 {
		return fmt.Errorf("--iteration must be positive integer, input: %v", flag.Trials)
	}

	return nil
}

func checkData(cmd, name string) error {
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

	var reason = new(string)
	if !isDataNamePassed(cmd, name, reason) {
		return fmt.Errorf("%v", *reason)
	}

	return nil
}

func isDataNamePassed(cmd, name string, reason *string) bool {
	err := config.InitWorkDir()
	if err != nil {
		*reason = err.Error()
		return false
	}

	command := ""
	filePath := ""
	command = "keentune param tune --job"
	filePath = config.GetDumpPath("tuning_jobs.csv")

	if !file.IsPathExist(filePath) {
		*reason = fmt.Sprintf("The file '%v' path does not exist", filePath)
		return false
	}

	if file.HasRecord(filePath, "name", name) {
		runningJob := file.GetRecord(filePath, "status", "finish", "name")
		if runningJob != "" {
			return true
		}
	} else {
		*reason = fmt.Sprintf("the specified name '%v' not exists. Run [%v %v --iteration xxx] or specify a new name and try again", name, command, name)
		return false

	}
	return false
}

func checkTuningFlags(cmdName string, flag *TuneFlag) error {
	var err error
	jobFlag := "--data"
	if cmdName == "tune" {
		jobFlag = "--job"
	}

	if err = checkJob(cmdName, flag.Name); err != nil {
		return fmt.Errorf("%v %v", jobFlag, err)
	}

	if flag.Round <= 0 {
		return fmt.Errorf("--iteration must be positive integer, input: %v", flag.Round)
	}

	return nil
}

func checkJob(cmd, name string) error {
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

	var reason = new(string)
	if !isNamePassed(cmd, name, reason) {
		return fmt.Errorf("%v", *reason)
	}

	return nil
}

func isNamePassed(cmd, name string, reason *string) bool {
	err := config.InitWorkDir()
	if err != nil {
		*reason = err.Error()
		return false
	}

	command := ""
	filePath := ""
	if cmd == "tune" {
		command = "keentune param delete --job"
		filePath = config.GetDumpPath("tuning_jobs.csv")
	}

	if cmd == "sensitize" {
		command = "keentune sensitize delete --data"
		filePath = config.GetDumpPath("sensitize_jobs.csv")
	}

	if !file.IsPathExist(filePath) {
		*reason = fmt.Sprintf("The file '%v' path does not exist", filePath)
		return false
	}

	if file.HasRecord(filePath, "name", name) {
		*reason = fmt.Sprintf("the specified name '%v' already exists. Run [%v %v] or specify a new name and try again", name, command, name)
		return false
	}

	runningJob := file.GetRecord(filePath, "status", "running", "name")
	if runningJob != "" {
		*reason = fmt.Sprintf("Job %v is running, you can wait for finishing it or stop it.", runningJob)
		return false
	}

	return true
}
