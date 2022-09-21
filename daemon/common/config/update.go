package config

import (
	"fmt"
	"strings"
	"sync"

	"github.com/go-ini/ini"
)

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

