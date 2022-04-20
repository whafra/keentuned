package config

import (
	"fmt"
	"keentune/daemon/common/file"

	"github.com/go-ini/ini"
)

var BackupConf *KeentunedConf

func Update(fileName, jobName string) error {
	BackupConf = new(KeentunedConf)
	err := file.DeepCopy(BackupConf, KeenTune)
	if err != nil {
		return fmt.Errorf("backup keentune conf %v", err)
	}

	cfg, err := ini.Load(fileName)
	if err != nil {
		return fmt.Errorf("failed to parse %s, %v", fileName, err)
	}

	var targetGroupNames, benchGroupNames []string
	if hasGroupSections(cfg, targetGroupNames, TargetSectionPrefix) {
		KeenTune.Target = Target{}
		if err = KeenTune.getTargetGroup(cfg); err != nil {
			return err
		}
	}

	if hasGroupSections(cfg, benchGroupNames, BenchSectionPrefix) {
		KeenTune.BenchGroup = []BenchGroup{}
		if err = KeenTune.getBenchGroup(cfg, false); err != nil {
			return err
		}

		if KeenTune.DestIP == "" {
			KeenTune.DestIP = BackupConf.DestIP
		}

		if KeenTune.BenchConf == "" {
			KeenTune.BenchConf = BackupConf.BenchConf
		}
	}

	jobFile := fmt.Sprintf("%v/%v/keentuned.conf", GetTuningPath(""), jobName)

	return cfg.SaveTo(jobFile)
}
