package sensitize

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	m "keentune/daemon/modules"
	"os"
)

type SensiRecord struct {
	name      string
	starttime string
	endtime   string
	trials    string
	status    string
	algorithm string
}

//  table header const

const (
	NA     = "-"
	Format = "2006-01-02 15:04:05"

	// status
	Run    = "running"
	Stop   = "abort"
	Finish = "finish"
	Err    = "error"
)

// Jobs run sensitize jobs service
func (s *Service) Jobs(flag string, reply *string) error {
	defer func() {
		*reply = log.ClientLogMap[log.SensitizeJobs]
		log.ClearCliLog(log.SensitizeJobs)
	}()

	filepath := config.GetDumpPath(config.SensitizeCsv)
	content, err := ioutil.ReadFile(filepath)
	if string(content) == "" {
		log.Infof(log.SensitizeJobs, "No job found")
		return nil
	}

	file, err := os.Open(filepath)
	if err != nil {
		log.Errorf(log.SensitizeJobs, "Can not open the file, err: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	r := csv.NewReader(file)
	var SensiData []SensiRecord
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		var sensirecord SensiRecord
		for index, value := range record {
			switch index {
			case m.TrainNameIdx:
				sensirecord.name = value
			case m.TrainStartIdx:
				sensirecord.starttime = value
			case m.TrainEndIdx:
				sensirecord.endtime = value
			case m.TrainTrialsIdx:
				sensirecord.trials = value
			case m.TrainStatusIdx:
				sensirecord.status = value
			case m.TrainAlgoIdx:
				sensirecord.algorithm = value
			}
		}
		SensiData = append(SensiData, sensirecord)
	}
	for _, v := range SensiData {
		var listInfo string
		listInfo += fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v", v.name, v.algorithm, v.trials, v.status, v.starttime, v.endtime)
		log.Infof(log.SensitizeJobs, "%v", listInfo)
	}

	return nil
}

