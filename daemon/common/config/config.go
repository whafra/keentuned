package config

import (
	"fmt"
	"github.com/go-ini/ini"
	"keentune/daemon/common/file"
	"keentune/daemon/common/utils"
	"os"
	"strings"
)

// KeentunedConf
type KeentunedConf struct {
	Home          string
	Port          string
	BenchIP       string
	BenchPort     string
	BaseRound     int
	ExecRound     int
	AfterRound    int
	TargetIP      []string
	TargetPort    string
	BrainIP       string
	BrainPort     string
	Algorithm     string
	HeartbeatTime int
	DumpConf
	Sensitize
	LogConf
}

type DumpConf struct {
	BaseDump bool
	ExecDump bool
	BestDump bool
	DumpHome string
}

type Sensitize struct {
	Algorithm  string
	BenchRound int
	ResultDir  string
}

type LogConf struct {
	LogFileLvl  string
	FileName    string
	Interval    int
	BackupCount int
}

const (
	keentuneConfigFile = "/etc/keentune/conf/keentuned.conf"
)

var (
	ProgramNeedExit = make(chan bool, 1)

	//  ApplyResultChan receive apply result
	ApplyResultChan     []chan []byte
	ServeFinish         = make(chan bool, 1)
	BenchmarkResultChan = make(chan []byte, 1)
	SensitizeResultChan = make(chan []byte, 1)
)

var (
	// KeenTune ...
	KeenTune KeentunedConf

	// ParamAllFile ...
	ParamAllFile = "parameter/sysctl.json"

	IsInnerBenchRequests     []bool
	IsInnerApplyRequests     []bool
	IsInnerSensitizeRequests []bool
)

func init() {
	conf := new(KeentunedConf)
	if err := conf.Save(); err != nil {
		fmt.Printf("init Keentuned conf err:%v\n", err)
		os.Exit(1)
	}

	initChanAndIPMap()
}

func initChanAndIPMap() {
	IsInnerBenchRequests = make([]bool, len(KeenTune.TargetIP)+2)
	IsInnerApplyRequests = make([]bool, len(KeenTune.TargetIP)+2)
	IsInnerSensitizeRequests = make([]bool, len(KeenTune.TargetIP)+2)
	ApplyResultChan = make([]chan []byte, len(KeenTune.TargetIP)+2)

	for index, _ := range KeenTune.TargetIP {
		ApplyResultChan[index+1] = make(chan []byte, 1)
	}
}

func (c *KeentunedConf) Save() error {
	cfg, err := ini.Load(keentuneConfigFile)
	if err != nil {
		return fmt.Errorf("failed to parse %s, %v", keentuneConfigFile, err)
	}

	keentune := cfg.Section("keentuned")
	c.Home = file.DecoratePath(keentune.Key("KEENTUNED_HOME").MustString("/etc/keentune"))
	c.Port = keentune.Key("PORT").MustString("9871")
	c.HeartbeatTime = keentune.Key("HEARTBEAT_TIME").MustInt(30)

	bench := cfg.Section("benchmark")
	c.BenchIP = bench.Key("BENCH_IP").MustString("")
	c.BenchPort = bench.Key("BENCH_PORT").MustString("9874")
	c.BaseRound = bench.Key("BASELINE_BENCH_ROUND").MustInt(5)
	c.ExecRound = bench.Key("TUNING_BENCH_ROUND").MustInt(3)
	c.AfterRound = bench.Key("RECHECK_BENCH_ROUND").MustInt(10)

	target := cfg.Section("target")
	c.TargetIP, err = changeStringToSlice(target.Key("TARGET_IP").MustString(""))
	if err != nil {
		fmt.Printf("%v keentune check target ip %v", utils.ColorString("red", "[ERROR]"), err)
	}

	c.TargetPort = target.Key("TARGET_PORT").MustString("9873")

	brain := cfg.Section("brain")
	c.BrainIP = brain.Key("BRAIN_IP").MustString("")
	c.BrainPort = brain.Key("BRAIN_PORT").MustString("9872")
	c.Algorithm = brain.Key("ALGORITHM").MustString("tpe")

	dump := cfg.Section("dump")
	c.DumpConf.BaseDump = dump.Key("DUMP_BASELINE_CONFIGURATION").MustBool(false)
	c.DumpConf.ExecDump = dump.Key("DUMP_TUNING_CONFIGURATION").MustBool(false)
	c.DumpConf.BestDump = dump.Key("DUMP_BEST_CONFIGURATION").MustBool(false)
	c.DumpConf.DumpHome = dump.Key("DUMP_HOME").MustString("")

	sensitize := cfg.Section("sensitize")
	c.Sensitize.Algorithm = sensitize.Key("ALGORITHM").MustString("random")
	c.Sensitize.BenchRound = sensitize.Key("BENCH_ROUND").MustInt(2)

	c.GetLogConf(cfg)

	KeenTune = *c
	return nil
}

func (c *KeentunedConf) GetLogConf(cfg *ini.File) {
	logInst := cfg.Section("log")
	c.LogConf.LogFileLvl = logInst.Key("LOGFILE_LEVEL").MustString("DEBUG")
	c.LogConf.FileName = logInst.Key("LOGFILE_NAME").MustString("keentuned.log")
	c.LogConf.Interval = logInst.Key("LOGFILE_INTERVAL").MustInt(2)
	c.LogConf.BackupCount = logInst.Key("LOGFILE_BACKUP_COUNT").MustInt(14)
}

func changeStringToSlice(ipString string) ([]string, error) {
	validIPs, invalidIPs := utils.CheckIPValidity(strings.Split(ipString, ","))
	if len(invalidIPs) != 0 {
		return validIPs, fmt.Errorf("find invalid or repeated ip %v, please check and restart", invalidIPs)
	}

	if len(validIPs) == 0 {
		return nil, fmt.Errorf("find valid ip is null, invalid ip is %v, please check and restart", invalidIPs)
	}

	return validIPs, nil
}

