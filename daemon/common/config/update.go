package config

import (
	"fmt"
	"keentune/daemon/common/file"
	"os"
	"strings"

	"github.com/go-ini/ini"
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

	return newCfg.SaveTo(jobFile)
}

func Backup(fileName, jobName string) error {
	defer func() {
		if file.IsPathExist(TuneTempConf) {
			os.Remove(TuneTempConf)
		}
	}()

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

func BackupSensitize(fileName, jobName string) error {
	defer func() {
		if file.IsPathExist(SensitizeTempConf) {
			os.Remove(SensitizeTempConf)
		}
	}()

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
