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
	Home       string
	Port       string
	BenchIP    string
	BenchPort  string
	BaseRound  int
	ExecRound  int
	AfterRound int
	TargetIP   []string
	TargetPort string
	Target
	BrainIP       string
	BrainPort     string
	Algorithm     string
	HeartbeatTime int
	DumpConf
	Sensitize
	LogConf
}

type Group struct {
	ParamMap map[string]map[string]interface{}
	IPs      []string
	Port     string
}

type Target struct {
	Group    []Group
	IPs      []string
	IPMap    map[string]int
	GroupMap map[string]int
	Ports    []string
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
	ConsoLvl    string
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
	KeenTune *KeentunedConf

	// ParamAllFile ...
	ParamAllFile = "parameter/param_all.json"

	IsInnerBenchRequests     []bool
	IsInnerApplyRequests     []bool
	IsInnerSensitizeRequests []bool
)

var RealLocalIP string

func init() {
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

	if err = c.getTargetGroup(cfg); err != nil {
		return err
	}

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
	return nil
}

func (c *KeentunedConf) getTargetGroup(cfg *ini.File) error {
	var groupNames []string
	sections := cfg.SectionStrings()
	for _, section := range sections {
		if strings.Contains(section, "target-group") {
			groupNames = append(groupNames, section)
		}
	}

	if len(groupNames) == 0 {
		return fmt.Errorf("target-group is null, please configure first")
	}

	var err error
	var ipDomain = make(map[string]map[string]string)
	var ipExist = make(map[string]bool)
	var id = new(int)
	c.Target.IPMap = make(map[string]int)
	for index, groupName := range groupNames {
		target := cfg.Section(groupName)
		var group Group
		group.IPs, err = changeStringToSlice(target.Key("TARGET_IP").MustString(""))
		if err != nil {
			return fmt.Errorf("keentune check target ip %v", err)
		}

		group.Port = target.Key("TARGET_PORT").MustString("9873")

		paramFiles := strings.Split(target.Key("PARAMETER").MustString(""), ",")
		var domains map[string]string
		domains, group.ParamMap, err = checkParamConf(paramFiles)
		if err != nil {
			return err
		}

		if err = IPDomainRepeated(groupName, group.IPs, domains, ipDomain); err != nil {
			return fmt.Errorf("%v", err)
		}

		c.Target.Group = append(c.Target.Group, group)
		c.Target.GroupMap[strings.TrimPrefix(groupName, "target-")] = index
		c.addIPMap(group.Port, group.IPs, ipExist, id)
	}

	return nil
}

func IPDomainRepeated(groupName string, ips []string, domains map[string]string, ipDomains map[string]map[string]string) error {
	for _, ip := range ips {
		_, exist := ipDomains[ip]
		if !exist {
			ipDomains[ip] = make(map[string]string)
		}

		for _, domain := range domains {
			originGroup, exist := ipDomains[ip][domain]
			if exist {
				return fmt.Errorf("Duplicate definition of parameter domain '%v'  in %v and %v!", domain, originGroup, groupName)
			}
			ipDomains[ip][domain] = groupName
		}
	}

	return nil
}

func (c *KeentunedConf) GetLogConf(cfg *ini.File) {
	logInst := cfg.Section("log")
	c.LogConf.ConsoLvl = logInst.Key("CONSOLE_LEVEL").MustString("INFO")
	c.LogConf.LogFileLvl = logInst.Key("LOGFILE_LEVEL").MustString("DEBUG")
	c.LogConf.FileName = logInst.Key("LOGFILE_NAME").MustString("keentuned.log")
	c.LogConf.Interval = logInst.Key("LOGFILE_INTERVAL").MustInt(2)
	c.LogConf.BackupCount = logInst.Key("LOGFILE_BACKUP_COUNT").MustInt(14)
}

func (c *KeentunedConf) addIPMap(port string, ips []string, ipExist map[string]bool, id *int) {
	for _, ip := range ips {
		if !ipExist[ip] {
			*id++
			ipExist[ip] = true
			c.Target.IPMap[ip] = *id
			c.Target.IPs = append(c.Target.IPs, ip)
			c.Target.Ports = append(c.Target.Ports, port)
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

