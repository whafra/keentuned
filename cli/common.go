package main

import (
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"os"
	"strings"
	"github.com/spf13/cobra"
)

var egBenchmark = "\tkeentune benchmark --job bench_test --bench benchmark/wrk/bench_wrk_nginx_long.json -i 10"

func subCommands() []*cobra.Command {
	var subCmds []*cobra.Command
	subCmds = append(subCmds, decorateCmd(createSensitizeCmds()))
	subCmds = append(subCmds, decorateCmd(createParamCmds()))
	subCmds = append(subCmds, decorateCmd(createProfileCmds()))
	subCmds = append(subCmds, decorateCmd(benchCmd()))
	subCmds = append(subCmds, decorateCmd(createRollbackAllCmd()))
	subCmds = append(subCmds, decorateCmd(initCmd()))

	return subCmds
}

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "init",
		Short:   "Initialize configuration",
		Long:    "Initialize configuration, Ping connectivity between nodes",
		Example: "\tkeentune init",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 0 {
				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
				cmd.Help()
				os.Exit(1)
			}

			RunInitRemote()
			return
		},
	}

	return cmd
}

func createRollbackAllCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rollbackall",
		Short:   "Restore all to initial state",
		Long:    "Restore all to initial state",
		Example: "\tkeentune rollbackall",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 0 {
				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
				cmd.Help()
				os.Exit(1)
			}

			RunRollbackAllRemote()
			return
		},
	}

	return cmd
}

func rollbackCmd(parentCmd string) *cobra.Command {
	var flag RollbackFlag
	cmd := &cobra.Command{
		Use:     "rollback",
		Short:   "Restore initial state",
		Long:    "Restore initial state",
		Example: fmt.Sprintf("\tkeentune %v rollback", parentCmd),
		Run: func(cmd *cobra.Command, args []string) {
			flag.Cmd = parentCmd
			RunRollbackRemote(cmd.Context(), flag)
			return
		},
	}

	return cmd
}

func setTuneFlag(cmd *cobra.Command, flag *TuneFlag) {
	flags := cmd.Flags()
	flags.StringVarP(&flag.Name, "job", "j", "", "Name of new knob auto-tuning job")
	flags.IntVarP(&flag.Round, "iteration", "i", 100, "MAX-iteration of knob auto-tuning")

	flags.StringVar(&flag.Config, "config", "keentuned.conf", "Customized config file for knob auto-tuning")

	flags.BoolVar(&flag.Verbose, "debug", false, "debug mode")
}

func stopCmd(flag string) *cobra.Command {
	var description string
	if flag == "param" {
		description = "Terminate a parameter tuning job"
	} else {
		description = "Terminate a sensitivity identification job"
	}
	var cmd = &cobra.Command{
		Use:     "stop",
		Short:   description,
		Long:    description,
		Example: fmt.Sprintf("\tkeentune %v stop", flag),
		Run: func(cmd *cobra.Command, args []string) {
			StopRemote(cmd.Context(), flag)
			return
		},
	}

	return cmd
}

func benchCmd() *cobra.Command {
	var flag BenchmarkFlag
	var cmd = &cobra.Command{
		Use:     "benchmark",
		Short:   "Automatic benchmark pressure test",
		Long:    "Automatic benchmark pressure test",
		Example: egBenchmark,
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Trim(flag.Name, " ") == "" || strings.Trim(flag.BenchConf, " ") == "" {
				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
				cmd.Usage()
				os.Exit(1)
			}

			RunBenchRemote(cmd.Context(), flag)
		},
		Hidden: true,
	}
	flags := cmd.Flags()
	flags.StringVarP(&flag.Name, "job", "j", "", "benchmark job name")
	flags.IntVarP(&flag.Round, "iteration", "i", 100, "benchmark execution iterations of pressure test")
	flags.StringVar(&flag.BenchConf, "bench", "", "benchmark configure infomation")
	return cmd
}

func newRootCmd() *cobra.Command {
        var isCatVersion bool
        cmd := &cobra.Command{
                Use:    "keentune [command]",
                Short:  "KeenTune is an AI tuning tool for Linux system and cloud applications",
                Long:   "KeenTune is an AI tuning tool for Linux system and cloud applications",
		Example: "\tkeentune init -h" +
                "\n\tkeentune param -h\n\tkeentune profile -h" +
                "\n\tkeentune rollbackall -h\n\tkeentune sensitize -h",
                RunE: func(cmd *cobra.Command, args []string) error {
                        if isCatVersion {
                                initWorkDirectory()
                                fmt.Printf("keentune version %v\n", config.KeenTune.VersionConf)
                                return nil
                        }

                        return cmd.Help()
                },
        }

        cmd.Flags().BoolVarP(&isCatVersion, "version", "v", false, "version message")
        return cmd
}

// confirm Interactive reply on terminal: [true] same as yes; false same as no.
func confirm() bool {
	var inputInfo string
	for {
		fmt.Scanln(&inputInfo)
		if inputInfo != "y" && inputInfo != "n" && inputInfo != "yes" && inputInfo != "no" && inputInfo != "Y" && inputInfo != "N" {
			fmt.Printf("\tplease input y(yes) or n(no)-->")
			continue
		}

		if inputInfo == "y" || inputInfo == "yes" || inputInfo == "Y" {
			return true
		}

		return false
	}
}

func decorateCmd(cmd *cobra.Command) *cobra.Command {
	var help bool
	cmd.Flags().BoolVarP(&help, "help", "h", false, "help message")
	return cmd
}

// ColorString print content string by color
func ColorString(color string, content string) string {
	// 其中0x1B是标记，[开始定义颜色，1代表高亮，40代表黑色背景，32代表绿色前景，0代表恢复默认颜色
	// 31 代表红色前景；33 代表黄色前景
	switch strings.ToUpper(color) {
	case "RED":
		return fmt.Sprintf("%c[1;40;31m%s%c[0m", 0x1B, content, 0x1B)
	case "GREEN":
		return fmt.Sprintf("%c[1;40;32m%s%c[0m", 0x1B, content, 0x1B)
	case "YELLOW":
		return fmt.Sprintf("%c[1;40;33m%s%c[0m", 0x1B, content, 0x1B)
	default:
		return content
	}
}

func IsTuningJobFinish(name string, status *string) bool {
	*status = file.GetRecord(config.GetDumpPath(config.TuneCsv), "name", name, "status")
	return *status == "finish"
}

