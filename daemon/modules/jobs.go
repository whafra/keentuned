package modules

import (
	"fmt"
	"io/ioutil"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"os"
)

// job type ...
const (
	// JobTuning job type is tuning
	JobTuning    = "tuning"
	JobProfile   = "profile"
	JobTraining  = "training"
	JobBenchmark = "benchmark"
)

// tuning job table header const
const (
	// TabName tuning table column: name
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

// TuneJobHeader ...
var TuneJobHeader = []string{
	TabName, TabAlgo, TabStatus, TabRound, TabCurRound,
	TabStart, TabEnd, TabCost, TabWSP, TabCmd, TabLog,
}

const activeJobCsv = "/var/keentune/runningJob.csv"

// train job table header const
const (
	// TabTrainName train table column: name
	TabTrainName   = "name"
	TabTrainStart  = "start_time"
	TabTrainEnd    = "end_time"
	TabTrainCost   = "total_time"
	TabTrainRound  = "trials"
	TabTrainStatus = "status"
	TabTrainLog    = "log"
	TabTrainWSP    = "workspace"
	TabTrainAlgo   = "algorithm"
	TabTrainPath   = "data_path"
)

// SensitizeJobHeader ...
var SensitizeJobHeader = []string{
	TabTrainName, TabTrainStart, TabTrainEnd, TabTrainCost, TabTrainRound,
	TabTrainStatus, TabTrainLog, TabTrainWSP, TabTrainAlgo,
	TabTrainPath,
}

func getTuneJobFile() string {
	return fmt.Sprint(config.GetDumpPath(config.TuneCsv))
}

func getSensitizeJobFile() string {
	return fmt.Sprint(config.GetDumpPath(config.SensitizeCsv))
}

// format and job status ...
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

// table column count
const (
	TuneCols  = 11
	TrainCols = 10
)

// tune job column index
const (
	// TuneNameIdx tuning index for name column
	TuneNameIdx = iota
	TuneAlgoIdx
	TuneStatusIdx
	TuneRoundIdx
	tuneCurRoundIdx
	TuneStartIdx
	TuneEndIdx
	tuneCostIdx
	TuneWSPIdx
	tuneCmdIdx
	tuneLogIdx
)

// train job column index
const (
	// TrainNameIdx training index of name column
	TrainNameIdx = iota
	TrainStartIdx
	TrainEndIdx
	trainCostIdx
	TrainTrialsIdx
	TrainStatusIdx
	trainLogIdx
	trainWSPIdx
	TrainAlgoIdx
	TrainDataIdx
)

// CreateTuneJob ...
func (tuner *Tuner) CreateTuneJob() error {
	cmd := fmt.Sprintf("keentune param tune --job %v -i %v", tuner.Name, tuner.MAXIteration)

	log := fmt.Sprintf("%v/%v.log", "/var/log/keentune", tuner.Name)

	jobInfo := []string{
		tuner.Name, tuner.Algorithm, Run, fmt.Sprint(tuner.MAXIteration),
		"0", tuner.StartTime.Format(Format), NA, NA,
		config.GetTuningPath(tuner.Name), cmd, log,
	}

	tuner.backupConfFile()
	return file.Insert(getTuneJobFile(), jobInfo)
}

func (tuner *Tuner) backupConfFile() {
	var workPath, filePath string
	if tuner.Flag == JobTuning {
		workPath = config.GetTuningPath(tuner.Name)
	}

	if tuner.Flag == JobTraining {
		workPath = config.GetSensitizePath(tuner.Job)
	}

	os.Mkdir(workPath, 0755)

	filePath = fmt.Sprintf("%v/keentuned.conf", workPath)
	file.Copy(config.GetKeenTunedConfPath(""), filePath)
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

// GetRunningTask ...
func GetRunningTask() string {
	file, _ := ioutil.ReadFile(activeJobCsv)
	return string(file)
}

// SetRunningTask ...
func SetRunningTask(class, name string) {
	content := fmt.Sprintf("%s %s", class, name)
	ioutil.WriteFile(activeJobCsv, []byte(content), 0666)
}

// ClearTask ...
func ClearTask() {
	os.Remove(activeJobCsv)
}

