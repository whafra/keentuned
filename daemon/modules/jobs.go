package modules

import (
	"fmt"
	"io/ioutil"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"os"
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

const activeJobCsv = "/var/keentune/runningJob.csv"

//  table header const
const (
	TabTrainName   = "name"
	TabTrainStart  = "start_time"
	TabTrainEnd    = "end_time"
	TabTrainCost   = "total_time"
	TabTrainRound  = "trials"
	TabTrainStatus = "status"
	TabTrainEpoch  = "epoch"
	TabTrainLog    = "log"
	TabTrainWSP    = "workspace"
	TabTrainAlgo   = "algorithm"
	TabTrainPath   = "data_path"
)

var SensitizeJobHeader = []string{
	TabTrainName, TabTrainStart, TabTrainEnd, TabTrainCost, TabTrainRound,
	TabTrainStatus, TabTrainEpoch, TabTrainLog, TabTrainWSP, TabTrainAlgo,
	TabTrainPath,
}

func getTuneJobFile() string {
	return fmt.Sprint(config.GetDumpPath("tuning_jobs.csv"))
}

func getSensitizeJobFile() string {
	return fmt.Sprint(config.GetDumpPath("sensitize_jobs.csv"))
}

const (
	NA     = "-"
	Format = "2006-01-02 15:04:05"

	// status
	Run    = "running"
	Stop   = "abort"
	Finish = "finish"
	Err    = "error"
	Kill   = "kill"
)

// tune job column index
const (
	TuneNameIdx = iota
	TuneAlgoIdx
	TuneStatusIdx
	TuneRoundIdx
	tuneCurRoundIdx
	TuneStartIdx
	TuneEndIdx
	tuneCostIdx
	tuneWSPIdx
	tuneCmdIdx
	tuneLogIdx
)

// tune job column index
const (
	TrainNameIdx = iota
	TrainStartIdx
	TrainEndIdx
	trainCostIdx
	TrainTrialsIdx
	TrainStatusIdx
	trainEpoch
	trainLogIdx
	trainWSPIdx
	TrainAlgoIdx
	trainDataPath
)

func (tuner *Tuner) CreateTuneJob() error {
	cmd := fmt.Sprintf("keentune param tune --job %v -i %v --config %v", tuner.Name, tuner.MAXIteration, config.GetCacheConfig())

	log := fmt.Sprintf("%v/%v.log", "/var/log/keentune", tuner.Name)

	jobInfo := []string{
		tuner.Name, tuner.Algorithm, Run, fmt.Sprint(tuner.MAXIteration),
		"0", tuner.StartTime.Format(Format), NA, NA,
		config.GetTuningPath(tuner.Name), cmd, log,
	}

	return file.Insert(getTuneJobFile(), jobInfo)
}

func (tuner *Tuner) updateJob(info map[int]interface{}) {
	var err error
	if tuner.Flag == "tuning" {
		err = file.UpdateRow(getTuneJobFile(), tuner.Name, info)
	} else if tuner.Flag == "training" {
		err = file.UpdateRow(getSensitizeJobFile(), tuner.Job, info)
	}

	if err != nil {
		log.Warnf("", "'%v' update '%v' %v", tuner.Flag, info, err)
		return
	}
}

func (tuner *Tuner) updateStatus(info string) {
	if tuner.Flag == "tuning" {
		tuner.updateJob(map[int]interface{}{TuneStatusIdx: info})
	} else if tuner.Flag == "training" {
		tuner.updateJob(map[int]interface{}{TrainStatusIdx: info})
	}
}

func GetRunningTask() string {
	file, _ := ioutil.ReadFile(activeJobCsv)
	return string(file)
}

func SetRunningTask(class, name string) {
	content := fmt.Sprintf("%s %s", class, name)
	ioutil.WriteFile(activeJobCsv, []byte(content), 0666)
}

func ClearTask() {
	os.Remove(activeJobCsv)
}

