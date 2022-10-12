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
	"sync"
)

type ymlTarget struct {
	IP        string   `yaml:"ip"`
	Available bool     `yaml:"available"`
	Knobs     []string `yaml:"knobs"`
	Domain    []string `yaml:"domain"`
}

type ymlBrain struct {
	BrainIP  string   `yaml:"ip"`
	AlgoTune []string `yaml:"algo_tuning"`
	AlgoSen  []string `yaml:"algo_sensi"`
}

type ymlBench struct {
	IP        string  `yaml:"ip"`
	Available bool    `yaml:"available"`
	Dest      ymlDest `yaml:"destination"`
	BenchConf string  `yaml:"benchmark"`
}

type ymlDest struct {
	IP        string `yaml:"ip"`
	Reachable bool   `yaml:"reachable"`
}

type keenTuneYML struct {
	Bench []ymlBench
	Brain ymlBrain
}

// Init ...
func (s *Service) Init(flag string, reply *string) error {
	result, err := initialize()
	if err != nil {
		log.Errorf("", "keentune init failed: %v", err)
		return err
	}

	if result != "" {
		*reply = fmt.Sprintf("%v  Connection failure Details:\n%v\n", utils.ColorString("yellow", "[Warning]"), result)
		log.Warnf("", "keentune init: %v", result)
		return nil
	}

	*reply = fmt.Sprintf("%v KeenTune Init success\n", utils.ColorString("green", "[OK]"))
	log.Info("", "KeenTune Init success")
	return nil
}

// initialize  KeenTune available test between brain, bench, target and daemon; Init Yaml create or update.
func initialize() (string, error) {
	err := config.CheckAndReloadConf()
	if err != nil {
		return "", err
	}

	var ymlConf = &keenTuneYML{}
	var (
		warningDetail = new(string)
		targetDetail  = new(string)
		benchDetail   = new(string)
		brainDetail   = new(string)
	)

	targetGroup := make([][]ymlTarget, len(config.KeenTune.Group))

	wg := sync.WaitGroup{}

	wg.Add(1)
	go checkBenchAVL(&wg, ymlConf, benchDetail)

	wg.Add(1)
	go checkBrainAVL(&wg, ymlConf, brainDetail)

	wg.Add(1)
	go checkTargetAVL(&wg, targetDetail, &targetGroup)

	wg.Wait()

	*warningDetail = *brainDetail
	*warningDetail += *benchDetail
	*warningDetail += *targetDetail

	bytes, err := yaml.Marshal(getYMLConf(ymlConf, targetGroup))
	if err != nil {
		return *warningDetail, err
	}

	err = ioutil.WriteFile(config.KeenTuneYMLFile, bytes, 0644)
	return strings.TrimSuffix(*warningDetail, "\n"), err
}

func checkBrainAVL(scWg *sync.WaitGroup, ymlConf *keenTuneYML, warningDetail *string) {
	defer scWg.Done()
	var err error

	ymlConf.Brain.BrainIP = config.KeenTune.BrainIP
	_, ymlConf.Brain.AlgoTune, ymlConf.Brain.AlgoSen, err = com.GetAVLDataAndAlgo()
	if err != nil {
		*warningDetail += fmt.Sprintf("\tbrain %v offline\n", config.KeenTune.BrainIP)
	}
}

func getYMLConf(conf *keenTuneYML, group [][]ymlTarget) interface{} {
	var ret = map[string]interface{}{}
	ret["brain"] = conf.Brain
	ret["bench-group-1"] = conf.Bench

	for idx, target := range config.KeenTune.Group {
		groupName := fmt.Sprintf("target-group-%v", target.GroupNo)
		ret[groupName] = group[idx]
	}

	return ret
}

func checkTargetAVL(scWg *sync.WaitGroup, warningDetail *string, targetGroup *[][]ymlTarget) {
	defer scWg.Done()
	accessResult := make([]string, len(config.KeenTune.IPMap))
	wg := sync.WaitGroup{}
	for groupIdx, target := range config.KeenTune.Group {
		(*targetGroup)[groupIdx] = make([]ymlTarget, len(target.IPs))
		for ipIdx, ip := range target.IPs {
			wg.Add(1)
			go accessTarget(&wg, accessResult, target, targetGroup, groupIdx, ip, ipIdx)
		}
	}

	wg.Wait()

	for _, detail := range accessResult {
		if len(detail) != 0 {
			*warningDetail += detail
		}
	}
}

func accessTarget(wg *sync.WaitGroup, warningDetail []string, target config.Group, targetGroup *[][]ymlTarget, groupIdx int, ip string, ipIdx int) {
	var tmpTarget ymlTarget
	var err error
	defer wg.Done()
	knobs := strings.Split(target.ParamConf, ",")
	for _, knob := range knobs {
		tmpTarget.Knobs = append(tmpTarget.Knobs, strings.TrimSpace(knob))
	}

	tmpTarget.Domain, err = com.GetAVLDomain(ip, target.Port)
	tmpTarget.IP = ip
	if err != nil {
		groupInfo := fmt.Sprintf("target-group[%v]:", target.GroupNo)
		idx := config.KeenTune.IPMap[ip] - 1
		warningDetail[idx] = fmt.Sprintf("\t%v %v offline\n", groupInfo, ip)
		(*targetGroup)[groupIdx][ipIdx] = tmpTarget
		return
	}

	tmpTarget.Available = true
	(*targetGroup)[groupIdx][ipIdx] = tmpTarget
}

func checkBenchAVL(scWg *sync.WaitGroup, ymlConf *keenTuneYML, warningDetail *string) {
	defer scWg.Done()
	ymlConf.Bench = make([]ymlBench, len(config.KeenTune.BenchIPMap))
	accessResult := make([]string, len(config.KeenTune.BenchIPMap))
	wg := sync.WaitGroup{}
	for _, bench := range config.KeenTune.BenchGroup {
		for _, ip := range bench.SrcIPs {
			wg.Add(1)
			go accessBench(ymlConf.Bench, &wg, bench, ip, accessResult)
		}
	}

	wg.Wait()

	for _, detail := range accessResult {
		if len(detail) != 0 {
			*warningDetail += detail
		}
	}
}

func accessBench(ymlBch []ymlBench, wg *sync.WaitGroup, bench config.BenchGroup, ip string, accessResult []string) {
	var tmpBench ymlBench
	var err error
	defer wg.Done()

	tmpBench.BenchConf = bench.BenchConf

	tmpBench.Available, tmpBench.Dest.Reachable, err = com.GetAVLAgentAddr(ip, bench.SrcPort, bench.DestIP)

	tmpBench.IP = ip
	tmpBench.Dest.IP = bench.DestIP

	ymlBch[config.KeenTune.BenchIPMap[ip]-1] = tmpBench
	if err != nil {
		idx := config.KeenTune.BenchIPMap[ip] - 1
		accessResult[idx] = fmt.Sprintf("%v", err)
	}
}

