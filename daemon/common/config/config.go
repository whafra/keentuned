package config

import (
	"fmt"
	"keentune/daemon/common/file"
	"keentune/daemon/common/utils"
	"os"
	"strconv"
	"strings"

	"github.com/go-ini/ini"
)

// KeentunedConf
type KeentunedConf struct {
	Default   `ini:"keentuned"`
	Bench     `ini:"benchmark"`
	Target    `ini:"-"`
	Brain     `ini:"brain"`
	Version   `ini:"version"`
	DumpConf  `ini:"dump"`
	Sensitize `ini:"sensitize"`
	LogConf   `ini:"log"`
}

type Version struct {
	VersionConf string `ini:"VERSION_NUM"`
}

type Brain struct {
	BrainIP   string `ini:"BRAIN_IP"`
	BrainPort string `ini:"BRAIN_PORT"`
	Algorithm string `ini:"ALGORITHM"`
}

type Default struct {
	Home          string `ini:"KEENTUNED_HOME"`
	Port          string `ini:"PORT"`
	HeartbeatTime int    `ini:"HEARTBEAT_TIME"`
}

type Bench struct {
	BenchGroup []BenchGroup   `ini:"-"`
	BaseRound  int            `ini:"BASELINE_BENCH_ROUND"`
	ExecRound  int            `ini:"TUNING_BENCH_ROUND"`
	AfterRound int            `ini:"RECHECK_BENCH_ROUND"`
	BenchConf  string         `ini:"BENCH_CONFIG"`
	BenchIPMap map[string]int `ini:"-"`
}

type BenchGroup struct {
	SrcIPs   []string
	SrcPort  string
	DestIP   string
	DestPort string
}

type Group struct {
	ParamMap  []DBLMap
	ParamConf string
	IPs       []string
	Port      string
	GroupName string //target-group-x
	GroupNo   int    //No. x of target-group-x
}

type Target struct {
	Group []Group
	IPMap map[string]int
}

type DumpConf struct {
	BaseDump bool   `ini:"DUMP_BASELINE_CONFIGURATION"`
	ExecDump bool   `ini:"DUMP_TUNING_CONFIGURATION"`
	BestDump bool   `ini:"DUMP_BEST_CONFIGURATION"`
	DumpHome string `ini:"DUMP_HOME"`
}

type Sensitize struct {
	Algorithm  string
	BenchRound int
}

type LogConf struct {
	LogFileLvl  string `ini:"LOGFILE_LEVEL"`
	FileName    string `ini:"LOGFILE_NAME"`
	Interval    int    `ini:"LOGFILE_INTERVAL"`
	BackupCount int    `ini:"LOGFILE_BACKUP_COUNT"`
}

// DBLMap Double Map
type DBLMap = map[string]map[string]interface{}

const (
	keentuneConfigFile = "/etc/keentune/conf/keentuned.conf"
)

var (
	ProgramNeedExit = make(chan bool, 1)

	//  ApplyResultChan receive apply result
	ApplyResultChan     []chan []byte
	ServeFinish         = make(chan bool, 1)
	BenchmarkResultChan []chan []byte
	SensitizeResultChan = make(chan []byte, 1)
)

var (
	// KeenTune ...
	KeenTune *KeentunedConf

	// ParamAllFile ...
	ParamAllFile = "parameter/sysctl.json"

	IsInnerBenchRequests     []bool
	IsInnerApplyRequests     []bool
	IsInnerSensitizeRequests []bool
)

var RealLocalIP string

const (
	TargetSectionPrefix = "target-group"
	BenchSectionPrefix  = "bench-group"
)

func Init() {
	KeenTune = new(KeentunedConf)
	err := KeenTune.Save()
	if err != nil {
		fmt.Printf("%v init Keentuned conf: %v\n", utils.ColorString("red", "[ERROR]"), err)
		os.Exit(1)
	}

	RealLocalIP, err = utils.GetExternalIP()
	if err != nil || RealLocalIP == "" {
		fmt.Printf("%v init Keentuned real local IP %v,err: %v\n", utils.ColorString("red", "[ERROR]"), RealLocalIP, err)
		os.Exit(1)
	}

	initChanAndIPMap()
}

func initChanAndIPMap() {
	IsInnerBenchRequests = make([]bool, len(KeenTune.BenchIPMap)+2)
	IsInnerApplyRequests = make([]bool, len(KeenTune.IPMap)+2)
	IsInnerSensitizeRequests = make([]bool, len(KeenTune.IPMap)+2)
	ApplyResultChan = make([]chan []byte, len(KeenTune.IPMap)+2)

	for _, index := range KeenTune.IPMap {
		ApplyResultChan[index] = make(chan []byte, 1)
	}

	BenchmarkResultChan = make([]chan []byte, len(KeenTune.BenchIPMap)+2)
	for _, benchIP := range KeenTune.BenchIPMap {
		BenchmarkResultChan[benchIP] = make(chan []byte, 1)
	}
}

func (c *KeentunedConf) Save() error {
	cfg, err := ini.InsensitiveLoad(keentuneConfigFile)
	if err != nil {
		return fmt.Errorf("failed to parse %s, %v", keentuneConfigFile, err)
	}

	keentune := cfg.Section("keentuned")
	c.Home = file.DecoratePath(keentune.Key("KEENTUNED_HOME").MustString("/etc/keentune"))
	c.Port = keentune.Key("PORT").MustString("9871")
	c.HeartbeatTime = keentune.Key("HEARTBEAT_TIME").MustInt(30)

	if err = c.getTargetGroup(cfg); err != nil {
		return err
	}

	if err = c.getBenchGroup(cfg); err != nil {
		return err
	}

	brain := cfg.Section("brain")
	c.BrainIP = brain.Key("BRAIN_IP").MustString("")
	c.BrainPort = brain.Key("BRAIN_PORT").MustString("9872")
	c.Brain.Algorithm = brain.Key("ALGORITHM").MustString("tpe")

	dump := cfg.Section("dump")
	c.DumpConf.BaseDump = dump.Key("DUMP_BASELINE_CONFIGURATION").MustBool(false)
	c.DumpConf.ExecDump = dump.Key("DUMP_TUNING_CONFIGURATION").MustBool(false)
	c.DumpConf.BestDump = dump.Key("DUMP_BEST_CONFIGURATION").MustBool(false)
	c.DumpConf.DumpHome = dump.Key("DUMP_HOME").MustString("")

	sensitize := cfg.Section("sensitize")
	c.Sensitize.Algorithm = sensitize.Key("ALGORITHM").MustString("random")
	c.Sensitize.BenchRound = sensitize.Key("BENCH_ROUND").MustInt(2)

	c.GetLogConf(cfg)

	version := cfg.Section("version")
	c.VersionConf = version.Key("VERSION_NUM").MustString("")

	return nil
}

func (c *KeentunedConf) getTargetGroup(cfg *ini.File) error {
	var groupNames = make([]string, 0)
	if !hasGroupSections(cfg, &groupNames, TargetSectionPrefix) {
		return fmt.Errorf("target-group is null, please configure first")
	}

	var err error
	var allGroupIPs = make(map[string]string)
	var ipExist = make(map[string]bool)
	var id = new(int)
	c.Target.IPMap = make(map[string]int)
	for _, groupName := range groupNames {
		target := cfg.Section(groupName)
		var group Group
		ipString := target.Key("TARGET_IP").MustString("localhost")
		group.IPs, err = changeStringToSlice(ipString)
		if err != nil {
			return fmt.Errorf("keentune check target ip %v", err)
		}

		group.Port = target.Key("TARGET_PORT").MustString("9873")

		group.GroupName = groupName
		groupName = groupName[13:] //截取“target-group-”后面的内容
		groupNo, err := strconv.Atoi(groupName)
		if err != nil || groupNo <= 0 {
			return fmt.Errorf("target-group is error, please check configure first")
		}
		group.GroupNo = groupNo
		group.ParamConf = target.Key("PARAMETER").MustString("sysctl.json")
		paramFiles := strings.Split(group.ParamConf, ",")

		_, group.ParamMap, err = checkParamConf(paramFiles)
		if err != nil {
			return err
		}

		if err = checkIPRepeated(groupName, group.IPs, allGroupIPs); err != nil {
			return fmt.Errorf("%v", err)
		}
		c.Target.Group = append(c.Target.Group, group)
		c.addTargetIPMap(group.IPs, ipExist, id)
	}

	return nil
}

func hasGroupSections(cfg *ini.File, groupNames *[]string, sectionPrefix string) bool {
	sections := cfg.SectionStrings()
	for _, section := range sections {
		if strings.Contains(section, sectionPrefix) {
			*groupNames = append(*groupNames, section)
		}
	}

	return len(*groupNames) != 0
}

func (c *KeentunedConf) getBenchGroup(cfg *ini.File) error {
	var groupNames = make([]string, 0)
	if !hasGroupSections(cfg, &groupNames, BenchSectionPrefix) {
		return fmt.Errorf("bench-group is null, please configure first")
	}

	var err error
	var allGroupIPs = make(map[string]string)
	var ipExist = make(map[string]bool)
	var id = new(int)
	c.Bench.BenchIPMap = make(map[string]int)
	for _, groupName := range groupNames {
		bench := cfg.Section(groupName)
		var group BenchGroup
		ipStringSrc := bench.Key("BENCH_SRC_IP").MustString("localhost")
		group.SrcIPs, err = changeStringToSlice(ipStringSrc)
		if err != nil {
			return fmt.Errorf("keentune check bench ip %v", err)
		}

		group.SrcPort = bench.Key("BENCH_SRC_PORT").MustString("9874")

		if err = checkIPRepeated(groupName, group.SrcIPs, allGroupIPs); err != nil {
			return fmt.Errorf("%v", err)
		}

		group.DestIP = bench.Key("BENCH_DEST_IP").MustString("localhost")
		group.DestPort = bench.Key("BENCH_DEST_PORT").MustString("9875")

		c.Bench.BenchGroup = append(c.Bench.BenchGroup, group)
		c.addBenchIPMap(group.SrcIPs, ipExist, id)
	}

	return c.getBenchmark(cfg)
}

func (c *KeentunedConf) getBenchmark(cfg *ini.File) error {
	bench := cfg.Section("benchmark")
	if bench == nil {
		return nil
	}

	c.BaseRound = bench.Key("BASELINE_BENCH_ROUND").MustInt(5)
	c.ExecRound = bench.Key("TUNING_BENCH_ROUND").MustInt(3)
	c.AfterRound = bench.Key("RECHECK_BENCH_ROUND").MustInt(10)
	c.BenchConf = bench.Key("BENCH_CONFIG").MustString("bench_wrk_nginx_long.json")

	return checkBenchConf(&c.BenchConf)
}

func checkIPRepeated(groupName string, ips []string, allGroupIPs map[string]string) error {
	for _, ip := range ips {
		_, exist := allGroupIPs[ip]
		if !exist {
			allGroupIPs[ip] = groupName
			continue
		}

		return fmt.Errorf("Duplicate ip '%v' in groups %v and %v!", ip, allGroupIPs[ip], groupName)
	}

	return nil
}

func (c *KeentunedConf) GetLogConf(cfg *ini.File) {
	logInst := cfg.Section("log")
	c.LogConf.LogFileLvl = logInst.Key("LOGFILE_LEVEL").MustString("DEBUG")
	c.LogConf.FileName = logInst.Key("LOGFILE_NAME").MustString("keentuned.log")
	c.LogConf.Interval = logInst.Key("LOGFILE_INTERVAL").MustInt(2)
	c.LogConf.BackupCount = logInst.Key("LOGFILE_BACKUP_COUNT").MustInt(14)
}

func (c *KeentunedConf) addTargetIPMap(ips []string, ipExist map[string]bool, id *int) {
	for _, ip := range ips {
		if !ipExist[ip] {
			*id++
			ipExist[ip] = true
			c.Target.IPMap[ip] = *id
		}
	}
}

func (c *KeentunedConf) addBenchIPMap(ips []string, ipExist map[string]bool, id *int) {
	for _, ip := range ips {
		if !ipExist[ip] {
			*id++
			ipExist[ip] = true
			c.Bench.BenchIPMap[ip] = *id
		}
	}
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

func InitWorkDir() error {
	cfg, err := ini.InsensitiveLoad(keentuneConfigFile)
	if err != nil {
		return fmt.Errorf("failed to parse %s, %v", keentuneConfigFile, err)
	}

	KeenTune = new(KeentunedConf)
	getWorkDir(cfg)
	return nil
}

func getWorkDir(cfg *ini.File) {
	keentune := cfg.Section("keentuned")
	KeenTune.Home = file.DecoratePath(keentune.Key("KEENTUNED_HOME").MustString("/etc/keentune"))

	dump := cfg.Section("dump")
	KeenTune.DumpConf.DumpHome = dump.Key("DUMP_HOME").MustString("")
}

func InitTargetGroup() error {
	cfg, err := ini.InsensitiveLoad(keentuneConfigFile)
	if err != nil {
		return fmt.Errorf("failed to parse %s, %v", keentuneConfigFile, err)
	}

	KeenTune = new(KeentunedConf)

	getWorkDir(cfg)
	KeenTune.GetLogConf(cfg)

	err = KeenTune.getTargetGroup(cfg)
	if err != nil {
		return fmt.Errorf("get target group %v", err)
	}

	return nil
}

func InitBrainConf() error {
	cfg, err := ini.InsensitiveLoad(keentuneConfigFile)
	if err != nil {
		return fmt.Errorf("failed to parse %s, %v", keentuneConfigFile, err)
	}

	KeenTune = new(KeentunedConf)

	brain := cfg.Section("brain")
	KeenTune.BrainIP = brain.Key("BRAIN_IP").MustString("")
	KeenTune.BrainPort = brain.Key("BRAIN_PORT").MustString("9872")
	KeenTune.Brain.Algorithm = brain.Key("ALGORITHM").MustString("tpe")

	getWorkDir(cfg)
	KeenTune.GetLogConf(cfg)
	return nil

}

