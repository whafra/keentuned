package param

import (
	"keentune/daemon/common/log"
	"encoding/csv"
	"os"
	"io/ioutil"
	"io"
	"fmt"
)

type CSVRecord struct {
	name string
	algorithm string
	iteration string
	status string
	starttime string
	endtime string
}

// Jobs run param jobs service
func (s *Service) Jobs(flag string, reply *string) error {
	defer func() {
		*reply = log.ClientLogMap[log.ParamJobs]
		log.ClearCliLog(log.ParamJobs)
	}()
	filepath := "/var/keentune/tuning_jobs.csv"
	content, err := ioutil.ReadFile(filepath)
	if string(content) == "" {
		log.Infof(log.ParamJobs,"No job found")
		return nil
	}

	file, err := os.Open(filepath)
	if err != nil {
		log.Errorf(log.ParamJobs, "Can not open the file, err: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()
	
	r := csv.NewReader(file)
	var csvData []CSVRecord
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		var csvrecord CSVRecord
		for index, value := range record {
			switch index {
			case 0:
				csvrecord.name = value
			case 1:
				csvrecord.algorithm = value
			case 2:
				csvrecord.iteration = value
			case 3:
				csvrecord.status = value
			case 4:
				csvrecord.starttime = value
			case 5:
				csvrecord.endtime = value
			}
		}
		csvData = append(csvData, csvrecord)
	}
	for _, v := range csvData {
		var listInfo string
		listInfo += fmt.Sprintf("%v\t%v\t%v\t%v\t%v\t%v",v.name, v.algorithm, v.iteration, v.status, v. starttime, v.endtime) 
	        log.Infof(log.ParamJobs,"%v", listInfo)

	}

	return nil
}
