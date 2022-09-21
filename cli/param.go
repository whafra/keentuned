package main

import (
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const (
	egTune          = "\tkeentune param tune --job tune_test --iteration 10\n\tkeentune param tune --job tune_test"
	egDump          = "\tkeentune param dump --job tune_test"
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
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
			cmd.Usage()
			os.Exit(1)
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
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Trim(flag.Name, " ") == "" {
				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
				cmd.Usage()
				os.Exit(1)
			}

			initWorkDirectory()
			if err := checkTuningFlags("tune", &flag); err != nil {
				fmt.Printf("%v check input: %v\n", ColorString("red", "[ERROR]"), err)
				os.Exit(1)
			}

			flag.Log = fmt.Sprintf("%v/%v.log", "/var/log/keentune", flag.Name)

			RunTuneRemote(flag)
		},
	}

	setTuneFlag(cmd, &flag)
	return cmd
}

func listParamCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List parameter and benchmark configuration files",
		Long:    "List parameter and benchmark configuration files",
		Example: egParamList,
		Run: func(cmd *cobra.Command, args []string) {
			RunListRemote("param")
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
			RunJobsRemote("param")
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
				cmd.Usage()
				os.Exit(1)
			}
			flag.Cmd = "param"

			initWorkDirectory()
			// Determine whether job already exists
			JobPath := config.GetTuningPath(flag.Name)
			_, err := os.Stat(JobPath)
			if err != nil {
				fmt.Printf("%v Auto-tuning job '%v' does not exist.\n", ColorString("red", "[ERROR]"), flag.Name)
				os.Exit(1)
			}
			// Determine whether job can be deleted
			if file.IsJobRunning(config.GetDumpPath(config.TuneCsv), flag.Name) {
				fmt.Printf("%v Auto-tuning job %v is running, use 'keentune param stop' to shutdown.\n", ColorString("yellow", "[Warning]"), flag.Name)
				return
			} else {
				fmt.Printf("%s %s '%s' ?Y(yes)/N(no)", ColorString("yellow", "[Warning]"), deleteTips, flag.Name)
				if !confirm() {
					fmt.Println("[-] Give Up Delete")
					return
				}
				RunDeleteRemote(flag)
				return
			}
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
				cmd.Usage()
				os.Exit(1)
			}

			err := checkDumpParam(&dump)
			if err != nil {
				fmt.Printf("%v Check dump param:%v\n", ColorString("red", "[ERROR]"), err)
				os.Exit(1)
			}

			RunDumpRemote(dump)
			return
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&dump.Name, "job", "j", "", "dynamic parameter tuning job name, query by command \"keentune param jobs\"")
	return cmd
}

func checkDumpParam(dump *DumpFlag) error {
	err := config.InitWorkDir()
	if err != nil {
		return fmt.Errorf("init work path %v", err)
	}

	status := new(string)
	if !IsTuningJobFinish(dump.Name, status) {
		return fmt.Errorf("job %v status is %v, dump is not supported", dump.Name, *status)
	}

	workPath := config.GetProfileWorkPath("")
	job := config.GetTuningPath(dump.Name)
	if !file.IsPathExist(job) {
		return fmt.Errorf("find the tuned file [%v] does not exist, please confirm that the tuning job [%v] exists or is completed. ", job, dump.Name)
	}

	const bestSuffix = "_best.json"
	_, bestFiles, err := file.WalkFilePath(job, bestSuffix)
	if err != nil {
		return fmt.Errorf("search the job '%v' file path err: %v ", dump.Name, err)
	}

	if len(bestFiles) == 0 {
		return fmt.Errorf("find the job '%v' best json doesn't exist", dump.Name)
	}

	var fileExist bool
	for _, bestJson := range bestFiles {
		parts := strings.Split(bestJson, bestSuffix)
		if len(parts) != 2 {
			return fmt.Errorf("best json name '%v' doesn't match 'jobName_group+id_best.josn'", bestJson)
		}

		combination := fmt.Sprintf("%s/%s,%s/%s.conf", job, bestJson, workPath, parts[0])
		dump.Output = append(dump.Output, combination)

		if !fileExist {
			fileName := fmt.Sprintf("%s/%s.conf", workPath, parts[0])
			fileExist = fileExist || file.IsPathExist(fileName)
		}
	}

	outputTips := "Dump %v has already operated, overwrite? Y(yes)/N(no)"
	if fileExist {
		fmt.Printf("%s %s", ColorString("yellow", "[Warning]"), fmt.Sprintf(outputTips, dump.Name))
		if !confirm() {
			return fmt.Errorf("outputFile exist and you have given up to overwrite it")
		}
	}

	return nil
}

