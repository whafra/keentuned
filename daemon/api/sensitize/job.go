package sensitize

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"os"
)

type SensiRecord struct {
	name      string
	algorithm string
	trails    string
	status    string
	starttime string
	endtime   string
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

// tune job column index
const (
	tuneNameIdx = iota
	tuneAlgoIdx
	tuneStatusIdx
	tuneRoundIdx
	tuneCurRoundIdx
	tuneStartIdx
	tuneEndIdx
	tuneCostIdx
	tuneWSPIdx
	tuneCmdIdx
	tuneLogIdx
)

// List run sensitize list service
func (s *Service) Jobs(flag string, reply *string) error {
	defer func() {
		*reply = log.ClientLogMap[log.SensitizeJobs]
		log.ClearCliLog(log.SensitizeJobs)
	}()

	filepath := "/var/keentune/sensitize_workspace.csv"
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
			case 0:
				sensirecord.name = value
			case 1:
				sensirecord.algorithm = value
			case 2:
				sensirecord.trails = value
			case 3:
				sensirecord.status = value
			case 4:
				sensirecord.starttime = value
			case 5:
				sensirecord.endtime = value
			}
		}
		SensiData = append(SensiData, sensirecord)
	}
	for _, v := range SensiData {
		var listInfo string
		listInfo += fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v", v.name, v.algorithm, v.trails, v.status, v.starttime, v.endtime)
		log.Infof(log.SensitizeJobs, "%v", listInfo)
	}

	return nil
}

func createTuneJob(flags TrainFlag) error {
	cmd := fmt.Sprintf("keentune sensitize train --data %v --job %v --trials %v --config %v", flags.Data, flags.Job, flags.Trials, flags.Config)

	log := fmt.Sprintf("%v/%v.log", "/var/log/keentune", flags.Job)

	jobInfo := []string{
		flags.Job, NA, NA, NA, fmt.Sprint(flags.Trials), Run,
		"0", log, config.GetSensitizeWorkPath(flags.Job), cmd, log,
	}
	return file.Insert(getSensitizeJobFile(), jobInfo)
}

func updateJob(flags TrainFlag, info map[int]interface{}) {
	var err error
	err = file.UpdateRow(getSensitizeJobFile(), flags.Job, info)
	if err != nil {
		log.Warnf("", "update '%v' %v", info, err)
		return
	}
}

func updateStatus(flags TrainFlag, info string) {
	updateJob(flags, map[int]interface{}{tuneStatusIdx: info})
}

func getSensitizeJobFile() string {
	return fmt.Sprint(config.GetDumpPath("sensitize_jobs.csv"))
}
