package param

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

type tuningRecord struct {
	name      string
	algorithm string
	iteration string
	status    string
	starttime string
	endtime   string
}

// Jobs run param jobs service
func (s *Service) Jobs(flag string, reply *string) error {
	defer func() {
		*reply = log.ClientLogMap[log.ParamJobs]
		log.ClearCliLog(log.ParamJobs)
	}()
	filepath := config.GetDumpPath(config.TuneCsv)
	content, err := ioutil.ReadFile(filepath)
	if string(content) == "" {
		log.Infof(log.ParamJobs, "No job found")
		return nil
	}

	file, err := os.Open(filepath)
	if err != nil {
		log.Errorf(log.ParamJobs, "Can not open the file, err: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	r := csv.NewReader(file)
	var tuningData []tuningRecord
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		var csvrecord tuningRecord
		for index, value := range record {
			switch index {
			case m.TuneNameIdx:
				csvrecord.name = value
			case m.TuneStartIdx:
				csvrecord.starttime = value
			case m.TuneEndIdx:
				csvrecord.endtime = value
			case m.TuneAlgoIdx:
				csvrecord.algorithm = value
			case m.TuneRoundIdx:
				csvrecord.iteration = value
			case m.TuneStatusIdx:
				csvrecord.status = value
			}
		}
		tuningData = append(tuningData, csvrecord)
	}
	
	for _, v := range tuningData {
		var listInfo string
		listInfo += fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v", v.name, v.algorithm, v.iteration, v.status, v.starttime, v.endtime)
		log.Infof(log.ParamJobs, "%v", listInfo)

	}

	return nil
}

