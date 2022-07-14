/*
Copyright © 2021 KeenTune

Package config for daemon, this package contains the check, config, priority, workpath. Parse the keentuned.conf file to get the ip, port information of the other three components.Determine whether the profile information meets and provide the file path related to keentune.
*/
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
	Default `ini:"keentuned"`
	Bench   `ini:"-"`
	Target  `ini:"-"`
	Brain   `ini:"brain"`
}

type Brain struct {
	BrainIP   string `ini:"BRAIN_IP"`
	BrainPort string `ini:"BRAIN_PORT"`
	Algorithm string `ini:"AUTO_TUNING_ALGORITHM"`
	Explainer string `ini:"SENSITIZE_ALGORITHM"`
}

type Default struct {
	Home          string `ini:"KEENTUNED_HOME"`
	Port          string `ini:"PORT"`
	HeartbeatTime int    `ini:"HEARTBEAT_TIME"`
	DumpHome      string `ini:"DUMP_HOME"`
	// dump control ...
	BaseDump bool `ini:"DUMP_BASELINE_CONFIGURATION"`
	ExecDump bool `ini:"DUMP_TUNING_CONFIGURATION"`
	BestDump bool `ini:"DUMP_BEST_CONFIGURATION"`

	// log ...
	LogFileLvl  string `ini:"LOGFILE_LEVEL"`
	FileName    string `ini:"LOGFILE_NAME"`
	Interval    int    `ini:"LOGFILE_INTERVAL"`
	BackupCount int    `ini:"LOGFILE_BACKUP_COUNT"`
	VersionConf string `ini:"VERSION_NUM"`

	// benchmark round ...
	BaseRound  int `ini:"BASELINE_BENCH_ROUND"`
	ExecRound  int `ini:"TUNING_BENCH_ROUND"`
	AfterRound int `ini:"RECHECK_BENCH_ROUND"`
}

type Bench struct {
	BenchGroup []BenchGroup   `ini:"-"`
	BenchIPMap map[string]int `ini:"-"`
}

type BenchGroup struct {
	SrcIPs    []string
	SrcPort   string
	DestIP    string
	BenchConf string
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
		RealLocalIP = "localhost"
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

	c.getDefault(cfg)

	if err = c.getTargetGroup(cfg); err != nil {
		return err
	}

	if err = c.getBenchGroup(cfg); err != nil {
		return err
	}

	brain := cfg.Section("brain")
	c.BrainIP = brain.Key("BRAIN_IP").MustString("")
	c.BrainPort = brain.Key("BRAIN_PORT").MustString("9872")
	c.Brain.Algorithm = brain.Key("AUTO_TUNING_ALGORITHM").MustString("tpe")

	c.Explainer = brain.Key("SENSITIZE_ALGORITHM").MustString("shap")

	return nil
}

func (c *KeentunedConf) getDefault(cfg *ini.File) {
	keentune := cfg.Section("keentuned")
	c.Home = file.DecoratePath(keentune.Key("KEENTUNED_HOME").MustString("/etc/keentune"))
	c.Port = keentune.Key("PORT").MustString("9871")
	c.HeartbeatTime = keentune.Key("HEARTBEAT_TIME").MustInt(30)

	c.BaseDump = keentune.Key("DUMP_BASELINE_CONFIGURATION").MustBool(false)
	c.ExecDump = keentune.Key("DUMP_TUNING_CONFIGURATION").MustBool(false)
	c.BestDump = keentune.Key("DUMP_BEST_CONFIGURATION").MustBool(false)
	c.DumpHome = keentune.Key("DUMP_HOME").MustString("")
	c.VersionConf = keentune.Key("VERSION_NUM").MustString("")

	c.BaseRound = keentune.Key("BASELINE_BENCH_ROUND").MustInt(1)
	c.ExecRound = keentune.Key("TUNING_BENCH_ROUND").MustInt(1)
	c.AfterRound = keentune.Key("RECHECK_BENCH_ROUND").MustInt(1)

	c.GetLogConf(cfg)
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
		group.BenchConf = bench.Key("BENCH_CONFIG").MustString("wrk_http_long.json")

		if err = checkBenchConf(&group.BenchConf); err != nil {
			return err
		}

		c.Bench.BenchGroup = append(c.Bench.BenchGroup, group)
		c.addBenchIPMap(group.SrcIPs, ipExist, id)
	}

	return nil
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
	logInst := cfg.Section("keentuned")
	c.LogFileLvl = logInst.Key("LOGFILE_LEVEL").MustString("DEBUG")
	c.FileName = logInst.Key("LOGFILE_NAME").MustString("keentuned.log")
	c.Interval = logInst.Key("LOGFILE_INTERVAL").MustInt(2)
	c.BackupCount = logInst.Key("LOGFILE_BACKUP_COUNT").MustInt(14)
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

	KeenTune.DumpHome = keentune.Key("DUMP_HOME").MustString("")
	KeenTune.VersionConf = keentune.Key("VERSION_NUM").MustString("")
}

func InitTargetGroup() error {
	cfg, err := ini.InsensitiveLoad(keentuneConfigFile)
	if err != nil {
		return fmt.Errorf("failed to parse %s, %v", keentuneConfigFile, err)
	}

	KeenTune = new(KeentunedConf)

	getWorkDir(cfg)

	err = KeenTune.getTargetGroup(cfg)
	if err != nil {
		return fmt.Errorf("get target group %v", err)
	}

	return nil
}

func GetJobParamConfig(job string) (string, string, error) {
	jobPath := GetTuningPath(job)
	if !file.IsPathExist(jobPath) {
		return "", "", fmt.Errorf("job '%v' does not exist", job)
	}

	confFile := fmt.Sprintf("%v/keentuned.conf", jobPath)
	cfg, err := ini.InsensitiveLoad(confFile)
	if err != nil {
		return "", "", err
	}

	var targetGroupNames = make([]string, 0)
	if !hasGroupSections(cfg, &targetGroupNames, TargetSectionPrefix) {
		return "", "", fmt.Errorf("target-group not found")
	}

	var parameterConf string
	for _, groupName := range targetGroupNames {
		target := cfg.Section(groupName)
		parameterConf += fmt.Sprintf("%v:", groupName)
		parameter := target.Key("PARAMETER").MustString("")
		confs := strings.Split(parameter, ",")
		for _, conf := range confs {
			fullPath := GetAbsolutePath(conf, "parameter", ".json", "_best.json")
			parameterConf += fmt.Sprintf(" %v,", fullPath)
		}
		parameterConf = fmt.Sprintf("%v\n", strings.TrimSuffix(parameterConf, ","))
	}

	var benchGroupNames = make([]string, 0)
	if !hasGroupSections(cfg, &benchGroupNames, BenchSectionPrefix) {
		return parameterConf, "", fmt.Errorf("bench-group not found")
	}

	var benchConf string
	for _, groupName := range benchGroupNames {
		bench := cfg.Section(groupName)
		benchName := bench.Key("BENCH_CONFIG").MustString("wrk_http_long.json")
		benchConf += GetBenchJsonPath(benchName)
	}

	return strings.TrimSuffix(parameterConf, "\n"), benchConf, nil
}

