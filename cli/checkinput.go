package main

import (
	"fmt"
	com "keentune/daemon/api/common"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"regexp"
)

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

	if cmd == "tune" && isTuneNameRepeat(name) {
		return fmt.Errorf("the specified name [%v] already exists. Run [keentune param delete --job %v] or specify a new name and try again", name, name)
	}

	if cmd == "sensitize" && com.IsDataNameUsed(name) {
		return fmt.Errorf("the specified name [%v] already exists. Run [keentune sensitize delete --data %v] or specify a new name and try again", name, name)
	}

	return nil
}

func isTuneNameRepeat(name string) bool {
	err := config.InitWorkDir()
	if err != nil {
		return false
	}

	tuneList, err := file.WalkFilePath(config.GetTuningWorkPath("")+"/", "", true, "/generate/")
	if err != nil {
		return false
	}

	for _, has := range tuneList {
		if has == name {
			return true
		}
	}

	return false
}

