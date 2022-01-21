package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"time"
)

const (
	egTune          = "\tkeentune param tune --param sysctl.json --bench bench_wrk_nginx_long.json --job tune_test --iteration 10\n\tkeentune param tune --param sysctl.json --bench bench_wrk_nginx_long.json --job tune_test"
	egDump          = "\tkeentune param dump --job tune_test --output tune_test.conf"
	egParamDel      = "\tkeentune param delete --job tune_test"
	egParamList     = "\tkeentune param list"
	egParamRollback = "\tkeentune param rollback"
	egJobs          = "\tkeentune param jobs"
	egParamStop     = "\tkeentune param stop"
)

func createParamCmds() *cobra.Command {
	var paramCmd = &cobra.Command{
		Use:     "param [command]",
		Short:   "Dynamic parameter tuning with AI algorithms",
		Long:    "Dynamic parameter tuning with AI algorithms",
		Example: fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s\n%s", egParamDel, egDump, egJobs, egParamList, egParamRollback, egParamStop, egTune),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				if args[0] != "--help" && args[0] != "-h" && args[0] != "tune" && args[0] != "list" && args[0] != "jobs" && args[0] != "delete" && args[0] != "dump" && args[0] != "rollback" && args[0] != "stop" {
					fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
				}
			}

			if len(args) == 0 {
				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
			}

			return cmd.Help()
		},
	}

	var paramCommands []*cobra.Command
	paramCommands = append(paramCommands, decorateCmd(tuneCmd()))
	paramCommands = append(paramCommands, decorateCmd(dumpCmd()))
	paramCommands = append(paramCommands, decorateCmd(listParamCmd()))
	paramCommands = append(paramCommands, decorateCmd(rollbackCmd("param")))
	paramCommands = append(paramCommands, decorateCmd(deleteParamJobCmd()))
	paramCommands = append(paramCommands, decorateCmd(stopCmd("param")))
	paramCommands = append(paramCommands, decorateCmd(jobCmd()))
	paramCmd.AddCommand(paramCommands...)

	return paramCmd
}

func tuneCmd() *cobra.Command {
	var flag TuneFlag
	cmd := &cobra.Command{
		Use:     "tune",
		Short:   "Deploy and start a parameter tuning job",
		Long:    "Deploy and start a parameter tuning job",
		Example: egTune,
		PreRun: func(cmd *cobra.Command, args []string) {
			if err := checkTuningFlags("tune", &flag); err != nil {
				fmt.Printf("%v check input: %v\n", ColorString("red", "[ERROR]"), err)
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Trim(flag.Name, " ") == "" || strings.Trim(flag.BenchConf, " ") == "" || strings.Trim(flag.ParamConf, " ") == "" {
				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
				cmd.Help()
				return
			}

			flag.Log = fmt.Sprintf("%v/%v-%v.log", "/var/log", "keentuned-param-tune", time.Now().Unix())

			RunTuneRemote(cmd.Context(), flag)
		},
	}

	setTuneFlag("tune", cmd, &flag)
	return cmd
}

func listParamCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List parameter and benchmark configuration files",
		Long:    "List parameter and benchmark configuration files",
		Example: egParamList,
		Run: func(cmd *cobra.Command, args []string) {
			RunListRemote(cmd.Context(), "param")
			return
		},
	}

	return cmd
}

func jobCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "jobs",
		Short:   "List parameter optimizing jobs",
		Long:    "List parameter optimizing jobs",
		Example: egJobs,
		Run: func(cmd *cobra.Command, args []string) {
			RunJobsRemote(cmd.Context())
			return
		},
	}

	return cmd
}

func deleteParamJobCmd() *cobra.Command {
	var flag DeleteFlag
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete the dynamic parameter tuning job",
		Long:    "Delete the dynamic parameter tuning job",
		Example: egParamDel,
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Trim(flag.Name, " ") == "" {
				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
				cmd.Help()
				return
			}

			flag.Cmd = "param"
			RunDeleteRemote(cmd.Context(), flag)
			return
		},
	}

	cmd.Flags().StringVarP(&flag.Name, "job", "j", "", "dynamic parameter tuning job name, query by command \"keentune param jobs\"")

	return cmd
}

func dumpCmd() *cobra.Command {
	var dump DumpFlag
	cmd := &cobra.Command{
		Use:     "dump",
		Short:   "Dump the parameter tuning result to a profile",
		Long:    "Dump the parameter tuning result to a profile",
		Example: egDump,
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Trim(dump.Name, " ") == "" {
				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
				cmd.Help()
				return
			}

			if strings.Trim(dump.Output, " ") == "" {
				dump.Output = dump.Name + ".conf"
			} else {
				dump.Output = strings.TrimSuffix(dump.Output, ".conf") + ".conf"
			}

			RunDumpRemote(cmd.Context(), dump)
			return
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&dump.Name, "job", "j", "", "dynamic parameter tuning job name, query by command \"keentune param jobs\"")
	flags.StringVarP(&dump.Output, "output", "o", "", "output profile file name, default with suffix \".conf\"")
	return cmd
}
