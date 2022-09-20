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

var (
	tuneConfig string
)

// ReSet ...
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
	cfg, err := ini.InsensitiveLoad(fileName)
	if err != nil {
		return fmt.Errorf("failed to parse %s, %v", fileName, err)
	}

	err = c.updateDefault(cfg, cmd)
	if err != nil {
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
	}

	if cmd == "training" {
		c.Explainer = algo
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
		sec.NewKey("BENCH_CONFIG", bench.BenchConf)
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

// Backup ...
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

	if fileName != TuneTempConf && cmd == "tuning" {
		SetCacheConfig(KeenTune.getConfigFlag())
	}

	return nil
}

func (c *KeentunedConf) getConfigFlag() string {
	var configFlag = "\""
	configFlag += fmt.Sprintf("ALGORITHM = %v\n", c.Brain.Algorithm)
	configFlag += fmt.Sprintf("BASELINE_BENCH_ROUND = %v\n", c.BaseRound)
	configFlag += fmt.Sprintf("TUNING_BENCH_ROUND = %v\n", c.ExecRound)
	configFlag += fmt.Sprintf("RECHECK_BENCH_ROUND = %v\n", c.AfterRound)
	configFlag += "\""
	return configFlag
}

// SetCacheConfig ...
func SetCacheConfig(info string) {
	tuneConfig = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(info, "\n", "\\n"), "\"", "'"), "''", "'")
}

// GetCacheConfig ...
func GetCacheConfig() string {
	var retConfig string
	retConfig = tuneConfig
	tuneConfig = ""

	return retConfig
}

func getBenchConf(benchGroup []BenchGroup) string {
	var benchConf string
	for index, bench := range benchGroup {
		benchConf += fmt.Sprintf("[bench-group-%v]\n", index+1)
		benchConf += fmt.Sprintf("BENCH_SRC_IP = %v\n", strings.Join(bench.SrcIPs, ","))
		benchConf += fmt.Sprintf("BENCH_SRC_PORT = %v\n", bench.SrcPort)
		benchConf += fmt.Sprintf("BENCH_DEST_IP = %v\n", bench.DestIP)
		benchConf += fmt.Sprintf("BENCH_CONFIG = %v\n", bench.BenchConf)
	}

	return benchConf
}

func getTargetConf(targetGroup []Group) string {
	var targetConf string
	for _, group := range targetGroup {
		targetConf += fmt.Sprintf("[target-group-%v]\n", group.GroupNo)
		targetConf += fmt.Sprintf("TARGET_IP = %v\n", strings.Join(group.IPs, ","))
		targetConf += fmt.Sprintf("TARGET_PORT = %v\n", group.Port)
		targetConf += fmt.Sprintf("PARAMETER = %v\n", group.ParamConf)
	}

	return targetConf
}

// UpdateKeentunedConf ...
func UpdateKeentunedConf(info string) (string, error) {
	details := strings.Split(info, "\n")
	if len(details) == 0 {
		return "", fmt.Errorf("info is empty")
	}

	var mutex = &sync.RWMutex{}
	mutex.Lock()
	defer mutex.Unlock()
	cfg, err := ini.Load(keentuneConfigFile)
	if err != nil {
		return "", fmt.Errorf("failed to parse %s, %v", "keentuned.conf", err)
	}

	var domain, result string
	for _, line := range details {
		pureLine := strings.TrimSpace(line)
		if len(pureLine) == 0 {
			continue
		}

		if strings.Contains(pureLine, "[") {
			domain = strings.Trim(strings.Trim(strings.TrimSpace(line), "]"), "[")
			continue
		}

		kvs := strings.Split(pureLine, "=")
		if len(kvs) != 2 {
			result += fmt.Sprintln(pureLine)
			continue
		}

		value := strings.TrimSpace(strings.Trim(kvs[1], "\""))
		cfg.Section(domain).Key(strings.TrimSpace(strings.ToUpper(kvs[0]))).SetValue(value)

	}

	if strings.Contains(info, "-group-") {
		err = checkInitGroup(cfg)
		if err != nil {
			return "", err
		}
	}

	err = cfg.SaveTo(keentuneConfigFile)
	if err != nil {
		return result, err
	}

	if result != "" {
		result = fmt.Sprintf("Warning partial success, failed configure as follows.\n %v", result)
		return result, nil
	}

	result = "keentuned configure save success"

	return result, nil
}

func checkInitGroup(cfg *ini.File) error {
	kdConf := new(KeentunedConf)
	if err := kdConf.getTargetGroup(cfg); err != nil {
		return err
	}

	if err := kdConf.getBenchGroup(cfg); err != nil {
		return err
	}

	return nil
}

// reloadConf ...
func reloadConf() error {
	var mutex = &sync.RWMutex{}
	mutex.Lock()
	defer mutex.Unlock()

	KeenTune = new(KeentunedConf)
	err := KeenTune.Save()
	if err != nil {
		return err
	}

	initChanAndIPMap()
	return nil
}

// CheckAndReloadConf check md5 of ENV keentuned.conf and reload conf to cache
func CheckAndReloadConf() error {
	md5Hash := GetKeenTuneConfFileMD5()
	if KeenTuneConfMD5 != md5Hash {
		KeenTuneConfMD5 = md5Hash
		err := reloadConf()
		if err != nil {
			return fmt.Errorf("reload conf failed: %v", err)
		}
	}

	return nil
}

// GetRerunConf get rerun configuration
func GetRerunConf(jobConf string) (map[string]interface{}, error) {
	cfg, err := ini.InsensitiveLoad(jobConf)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s, %v", keentuneConfigFile, err)
	}

	var retConf = make(map[string]interface{})
	kd := cfg.Section("keentuned")
	retConf["BaseRound"] = kd.Key("BASELINE_BENCH_ROUND").MustInt()
	retConf["TuningRound"] = kd.Key("TUNING_BENCH_ROUND").MustInt()
	retConf["RecheckRound"] = kd.Key("RECHECK_BENCH_ROUND").MustInt()

	return retConf, nil
}

