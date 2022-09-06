package system

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	com "keentune/daemon/api/common"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"keentune/daemon/common/utils"
	"strings"
)

type ymlTarget struct {
	IP     string   `yaml:"ip"`
	Knobs  []string `yaml:"knobs"`
	Domain []string `yaml:"domain"`
}

type ymlBrain struct {
	BrainIP  string   `yaml:"ip"`
	AlgoTune []string `yaml:"algo_tuning"`
	AlgoSen  []string `yaml:"algo_sensi"`
}

type ymlBench struct {
	IP        string `yaml:"ip"`
	Dest      string `yaml:"destination"`
	BenchConf string `yaml:"benchmark"`
}

type keenTuneYML struct {
	Bench  []ymlBench  `yaml:"bench"`
	Brain  ymlBrain    `yaml:"brain"`
	Hex    string      `yaml:"hex"`
	Target []ymlTarget `yaml:"target"`
}

// Init ...
func (s *Service) Init(flag string, reply *string) error {
	result, err := initialize()
	if err != nil {
		log.Errorf("", "keentune init failed: %v", err)
		return err
	}

	if result != "" {
		*reply = fmt.Sprintf("%v %v", utils.ColorString("yellow", "[Warning]"), result)
		log.Warnf("", "keentune init: %v", result)
		return nil
	}

	*reply = fmt.Sprintf("%v KeenTune Init success\n", utils.ColorString("green", "[OK]"))
	log.Info("", "KeenTune Init success")
	return nil
}

// initialize  KeenTune available test between brain, bench, target and daemon; Create or Update init.yaml file.
func initialize() (string, error) {
	var ymlConf = &keenTuneYML{}
	var warningDetail string
	var err error
	warningDetail = checkBenchAVL(ymlConf)

	ymlConf.Brain.BrainIP = config.KeenTune.BrainIP
	_, ymlConf.Brain.AlgoTune, ymlConf.Brain.AlgoSen, err = com.GetAVLDataAndAlgo()
	if err != nil {
		warningDetail += fmt.Sprintf("brain host %v unreachable\n", config.KeenTune.BrainIP)
	}

	targetResult := checkTargetAVL(ymlConf)
	warningDetail += targetResult

	bytes, err := yaml.Marshal(ymlConf)
	if err != nil {
		return warningDetail, err
	}

	err = ioutil.WriteFile(config.KeenTuneYMLFile, bytes, 0644)
	return strings.TrimSuffix(warningDetail, "\n"), err
}

func checkTargetAVL(ymlConf *keenTuneYML) string {
	var targetWarning string
	var err error
	ymlConf.Target = make([]ymlTarget, len(config.KeenTune.IPMap))
	for _, target := range config.KeenTune.Group {
		var tmpTarget ymlTarget
		for _, ip := range target.IPs {
			tmpTarget.Domain, err = com.GetAVLDomain(ip, target.Port)
			if err != nil {
				targetWarning += fmt.Sprintf("\ttarget host %v unreachable", ip)
				continue
			}

			tmpTarget.Knobs = strings.Split(target.ParamConf, ",")
			tmpTarget.IP = ip
			idx := config.KeenTune.IPMap[ip] - 1
			ymlConf.Target[idx] = tmpTarget
		}
	}

	return targetWarning
}

func checkBenchAVL(ymlConf *keenTuneYML) string {
	var warningDetails string
	ymlConf.Bench = make([]ymlBench, len(config.KeenTune.BenchIPMap))
	for _, bench := range config.KeenTune.BenchGroup {
		var tmpBench ymlBench
		tmpBench.BenchConf = bench.BenchConf
		tmpBench.Dest = bench.DestIP
		for _, ip := range bench.SrcIPs {
			err := utils.Ping(ip, bench.SrcPort)
			if err != nil {
				warningDetails += fmt.Sprintf("bench src host %v unreachable\n", ip)
				fmt.Printf("ip %v\n", warningDetails)
				continue
			}

			tmpBench.IP = ip
			ymlConf.Bench[config.KeenTune.BenchIPMap[ip]-1] = tmpBench
		}
	}

	return warningDetails
}

