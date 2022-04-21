package modules

import (
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
)

//  table header const
const (
	TabName     = "name"
	TabAlgo     = "algorithm"
	TabStatus   = "status"
	TabRound    = "iteration"
	TabCurRound = "current_iteration"
	TabStart    = "start_time"
	TabEnd      = "end_time"
	TabCost     = "total_time"
	TabWSP      = "workspace"
	TabCmd      = "cmd"
	TabLog      = "log"
)

var TuneJobHeader = []string{
	TabName, TabAlgo, TabStatus, TabRound, TabCurRound,
	TabStart, TabEnd, TabCost, TabWSP, TabCmd, TabLog,
}

var (
	tuneJobFile = fmt.Sprint(config.GetDumpPath("tuning_jobs.csv"))
)

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

func (tuner *Tuner) CreateTuneJob() error {
	cmd := fmt.Sprintf("keentune param tune --job %v -i %v", tuner.Name, tuner.MAXIteration)
	log := fmt.Sprintf("%v/%v.log", "/var/log/keentune", tuner.Name)

	jobInfo := []string{
		tuner.Name, tuner.Algorithm, Run, fmt.Sprint(tuner.MAXIteration),
		"0", tuner.StartTime.Format(Format), NA, NA,
		config.GetTuningWorkPath(tuner.Name), cmd, log,
	}

	return file.Insert(tuneJobFile, jobInfo)
}

func (tuner *Tuner) updateJob(info map[int]interface{}) {
	var err error
	if tuner.Flag == "tuning" {
		err = file.UpdateRow(tuneJobFile, tuner.Name, info)
	}

	if err != nil {
		log.Warnf("", "'%v' update '%v' %v", tuner.Flag, info, err)
		return
	}
}

func (tuner *Tuner) updateStatus(info string) {
	if tuner.Flag == "tuning" {
		tuner.updateJob(map[int]interface{}{tuneStatusIdx: info})
	}
}
