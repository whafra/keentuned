package config

import (
	"fmt"
	"io/ioutil"
	"keentune/daemon/common/file"
	"keentune/daemon/common/utils"
	"os"
	"strings"
	"sync"

	"github.com/go-ini/ini"
)

func ReSet() error {
	backupConf := new(KeentunedConf)
	err := backupConf.Save()
	if err != nil {
		return fmt.Errorf("reload Keentuned.conf: %v\n", err)
	}

	err = utils.DeepCopy(KeenTune, *backupConf)
	if err != nil {
		return fmt.Errorf("reset keentuned.conf: %v", err)
	}

	initChanAndIPMap()
	return nil
}

func update(fileName, cmd string) error {
	var err error
	if fileName == "" || fileName == keentuneConfigFile {
		KeenTune = new(KeentunedConf)
		err = KeenTune.Save()
		if err != nil {
			return fmt.Errorf("reload Keentuned.conf: %v\n", err)
		}

	} else {
		err = KeenTune.update(fileName, cmd)
		if err != nil {
			return fmt.Errorf("reload Keentuned.conf: %v\n", err)
		}
	}

	initChanAndIPMap()
	return nil
}

func (c *KeentunedConf) update(fileName, cmd string) error {
	backupConf := new(KeentunedConf)
	err := utils.DeepCopy(backupConf, c)
	if err != nil {
		return fmt.Errorf("deep copy: %v", err)
	}

	cfg, err := ini.InsensitiveLoad(fileName)
	if err != nil {
		return fmt.Errorf("failed to parse %s, %v", fileName, err)
	}

	err = c.updateDefault(cfg, cmd)
	if err != nil {
		return err
	}

	c.Target = Target{}
	if err = c.getTargetGroup(cfg); err != nil {
		return err
	}

	c.BenchGroup = []BenchGroup{}
	if err = c.getBenchGroup(cfg); err != nil {
		return err
	}

	return nil
}

func (c *KeentunedConf) updateDefault(cfg *ini.File, cmd string) error {
	empty := cfg.Section("")
	if empty == nil {
		return nil
	}

	// Required: algorithm
	algo := empty.Key("ALGORITHM").MustString("")
	if algo == "" {
		return fmt.Errorf("algorithm is required")
	}

	if cmd == "tuning" {
		c.Brain.Algorithm = algo
		// Required: baseline_bench_round, tuning_bench_round, recheck_bench_round
		c.BaseRound = empty.Key("BASELINE_BENCH_ROUND").MustInt(5)
		c.ExecRound = empty.Key("TUNING_BENCH_ROUND").MustInt(3)
		c.AfterRound = empty.Key("RECHECK_BENCH_ROUND").MustInt(10)

		// Optional: bench_config
		benchConf := empty.Key("BENCH_CONFIG").MustString("")
		if benchConf != "" {
			c.BenchConf = benchConf
			return checkBenchConf(&c.BenchConf)
		}
	}

	if cmd == "training" {
		c.Sensitize.Algorithm = algo
		// todo required: epoch ...
	}

	return nil
}

func dump(jobName string, cmd string) error {
	var jobFile string
	if cmd == "tuning" {
		jobFile = fmt.Sprintf("%v/%v/keentuned.conf", GetTuningPath(""), jobName)
	} else if cmd == "training" {
		jobFile = fmt.Sprintf("%v/%v/keentuned.conf", GetSensitizePath(""), jobName)
	}

	newCfg := ini.Empty()
	if err := ini.ReflectFrom(newCfg, KeenTune); err != nil {
		return err
	}

	for index, bench := range KeenTune.BenchGroup {
		sectionName := fmt.Sprintf("bench-group-%v", index+1)
		sec, err := newCfg.NewSection(sectionName)
		if err != nil {
			return fmt.Errorf("new bench group %v section %v", index+1, err)
		}

		sec.NewKey("BENCH_SRC_IP", strings.Join(bench.SrcIPs, ","))
		sec.NewKey("BENCH_SRC_PORT", bench.SrcPort)
		sec.NewKey("BENCH_DEST_IP", bench.DestIP)
		sec.NewKey("BENCH_DEST_PORT", bench.DestPort)
	}

	for index, target := range KeenTune.Target.Group {
		sectionName := fmt.Sprintf("target-group-%v", target.GroupNo)
		sec, err := newCfg.NewSection(sectionName)
		if err != nil {
			return fmt.Errorf("new target group %v section %v", index+1, err)
		}

		sec.NewKey("TARGET_IP", strings.Join(target.IPs, ","))
		sec.NewKey("TARGET_PORT", target.Port)
		sec.NewKey("PARAMETER", target.ParamConf)
	}

	if cmd == "tuning" {
		os.Mkdir(GetTuningPath(jobName), 0755)
	} else if cmd == "training" {
		os.Mkdir(GetSensitizePath(jobName), 0755)
	}

	dumpDefaultConfig(newCfg)
	return newCfg.SaveTo(jobFile)
}

func dumpDefaultConfig(cfg *ini.File) {
	if !file.IsPathExist(keentuneConfigFile + ".bak") {
		backup, _ := ioutil.ReadFile(keentuneConfigFile)
		ioutil.WriteFile(keentuneConfigFile+".bak", backup, 0644)
	}

	var mutex = &sync.RWMutex{}
	mutex.Lock()
	defer mutex.Unlock()
	err := cfg.SaveTo(keentuneConfigFile)
	if err != nil {
		fmt.Printf("%v tuning save config to default file err: %v\n", utils.ColorString("yellow", "[Warning]"), err)
	}
}

func Backup(fileName, jobName string, cmd string) error {
	var err error
	defer func() {
		if cmd == "tuning" {
			if file.IsPathExist(TuneTempConf) {
				os.Remove(TuneTempConf)
			}
		} else if cmd == "training" {
			if file.IsPathExist(SensitizeTempConf) {
				os.Remove(SensitizeTempConf)
			}
		}

		if err != nil {
			ReSet()
		}
	}()

	err = update(fileName, cmd)
	if err != nil {
		return fmt.Errorf("update %v", err)
	}

	err = dump(jobName, cmd)
	if err != nil {
		return fmt.Errorf("backup conf%v", err)
	}

	return nil
}

