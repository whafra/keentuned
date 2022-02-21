package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func subCommands() []*cobra.Command {
	var subCmds []*cobra.Command
	subCmds = append(subCmds, decorateCmd(createSensitizeCmds()))
	subCmds = append(subCmds, decorateCmd(createParamCmds()))
	subCmds = append(subCmds, decorateCmd(createProfileCmds()))
	subCmds = append(subCmds, decorateCmd(benchCmd()))

	return subCmds
}

var egBenchmark = "\tkeentune benchmark --job bench_test --bench benchmark/wrk/bench_wrk_nginx_long.json -i 10"

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

func setTuneFlag(cmdName string, cmd *cobra.Command, flag *TuneFlag) {
	flags := cmd.Flags()
	if cmdName == "tune" {
		flags.StringVarP(&flag.Name, "job", "j", "", "name of the new dynamic parameter tuning job")
		flags.IntVarP(&flag.Round, "iteration", "i", 100, "iteration of dynamic parameter tuning")
	} else {
		flags.StringVarP(&flag.Name, "data", "d", "", "sensitivity identification data name")
		flags.IntVarP(&flag.Round, "iteration", "i", 100, "iteration of sensitivity identification data collection")
	}

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
				cmd.Help()
				return
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

