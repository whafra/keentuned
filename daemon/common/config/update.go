package config

import (
	"fmt"
	"github.com/go-ini/ini"
	"keentune/daemon/common/file"
)

func ReSet() error {
	backupConf := new(KeentunedConf)
	err := backupConf.Save()
	if err != nil {
		return fmt.Errorf("reload Keentuned.conf: %v\n", err)
	}

	err = file.DeepCopy(KeenTune, backupConf)
	if err != nil {
		return fmt.Errorf("reset keentuned.conf: %v", err)
	}

	initChanAndIPMap()
	return nil
}

func update(fileName string) error {
	backupConf := new(KeentunedConf)
	err := file.DeepCopy(backupConf, KeenTune)
	if err != nil {
		return fmt.Errorf("deep copy: %v", err)
	}

	cfg, err := ini.InsensitiveLoad(fileName)
	if err != nil {
		return fmt.Errorf("failed to parse %s, %v", fileName, err)
	}

	empty := cfg.Section("")
	algo := empty.Key("ALGORITHM").MustString("")
	if algo != "" {
		KeenTune.Brain.Algorithm = algo
	}

	var targetGroupNames, benchGroupNames = make([]string, 0), make([]string, 0)
	if hasGroupSections(cfg, &targetGroupNames, TargetSectionPrefix) {
		KeenTune.Target = Target{}
		if err = KeenTune.getTargetGroup(cfg); err != nil {
			return err
		}
	}

	if hasGroupSections(cfg, &benchGroupNames, BenchSectionPrefix) {
		KeenTune.BenchGroup = []BenchGroup{}
		if err = KeenTune.getBenchGroup(cfg); err != nil {
			return err
		}
	}

	initChanAndIPMap()

	return nil
}

func dump(jobName string) error {
	jobFile := fmt.Sprintf("%v/%v/keentuned.conf", GetTuningPath(""), jobName)

	newCfg := ini.Empty()
	if err := ini.ReflectFrom(newCfg, KeenTune); err != nil {
		return err
	}

	return newCfg.SaveTo(jobFile)
}

func Backup(fileName, jobName string) error {
	err := update(fileName)
	if err != nil {
		return fmt.Errorf("update %v", err)
	}

	err = dump(jobName)
	if err != nil {
		return fmt.Errorf("backup conf%v", err)
	}

	return nil
}

