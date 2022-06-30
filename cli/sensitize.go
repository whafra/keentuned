package main

import (
	"fmt"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"keentune/daemon/common/log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const (
	egTrain         = "\tkeentune sensitize train --data collect_test --job train_test --trials 2"
	egDelete        = "\tkeentune sensitize delete --job collect_test"
	egSensitiveJobs = "\tkeentune sensitize jobs"
	egSensitiveStop = "\tkeentune sensitize stop"
)

func createSensitizeCmds() *cobra.Command {
	sensitizeCmd := &cobra.Command{
		Use:     "sensitize [command]",
		Short:   "Sensitive parameter identification and explanation with AI algorithms",
		Long:    "Sensitive parameter identification and explanation with AI algorithms",
		Example: fmt.Sprintf("%s\n%s\n%s\n%s", egDelete, egSensitiveJobs, egSensitiveStop, egTrain),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				if args[0] != "--help" && args[0] != "-h" /*&& args[0] != "collect" */ && args[0] != "jobs" && args[0] != "delete" && args[0] != "train" && args[0] != "stop" {
					fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
				}
			}

			if len(args) == 0 {
				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
			}

			return cmd.Help()
		},
	}

	var sesiCmds []*cobra.Command

	//sesiCmds = append(sesiCmds, decorateCmd(collectCmd()))
	sesiCmds = append(sesiCmds, decorateCmd(jobSensitivityCmd()))
	sesiCmds = append(sesiCmds, decorateCmd(trainCmd()))
	sesiCmds = append(sesiCmds, decorateCmd(deleteSensitivityCmd()))
	sesiCmds = append(sesiCmds, decorateCmd(stopCmd("sensitize")))

	sensitizeCmd.AddCommand(sesiCmds...)

	return sensitizeCmd
}

func trainCmd() *cobra.Command {
	var trainflags TrainFlag
	cmd := &cobra.Command{
		Use:     "train",
		Short:   "Deploy and start a sensitivity identification job",
		Long:    "Deploy and start a sensitivity identification job",
		Example: egTrain,
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Trim(trainflags.Data, " ") == "" {
				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
				cmd.Help()
				return
			}

			err := initSensitizeConf()
			if err != nil {
				fmt.Printf("%v Init Brain conf: %v\n", ColorString("red", "[ERROR]"), err)
				os.Exit(1)
			}			

			if strings.Trim(trainflags.Job, " ") == "" {
				trainflags.Job = trainflags.Data
			}

			if err := checkTrainingFlags("sensitize", &trainflags); err != nil {
				fmt.Printf("%v check input: %v\n", ColorString("red", "[ERROR]"), err)
				os.Exit(1)
			}

			trainflags.Log = fmt.Sprintf("%v/%v-%v.log", "/var/log/keentune", "keentuned-sensitize-train", trainflags.Job)
			RunTrainRemote(cmd.Context(), trainflags)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&trainflags.Data, "data", "d", "", "available sensitivity identification data, query by \"keentune sensitize jobs\"")
	flags.IntVarP(&trainflags.Trials, "trials", "t", 1, "sensitize trials, range [1,10]")
	flags.StringVarP(&trainflags.Job, "job", "j", "", "job file of sensitive parameter identification and explanation")
	flags.StringVar(&trainflags.Config, "config", "", "configuration specified for train")

	return cmd
}

func jobSensitivityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "jobs",
		Short:   "List available sensitivity identification jobs",
		Long:    "List available sensitivity identification jobs",
		Example: egSensitiveJobs,
		Run: func(cmd *cobra.Command, args []string) {
			RunJobsRemote(cmd.Context(), "sensitize")
			return
		},
	}

	return cmd
}

func deleteSensitivityCmd() *cobra.Command {
	var flag DeleteFlag
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete the sensitivity identification job",
		Long:    "Delete the sensitivity identification job",
		Example: egDelete,
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Trim(flag.Name, " ") == "" {
				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
				cmd.Help()
				return
			}
			flag.Cmd = "sensitize"
			err := initSensitizeConf()
			if err != nil {
				fmt.Printf("%v Init Brain conf: %v\n", ColorString("red", "[ERROR]"), err)
				os.Exit(1)
			}

			//Determine whether job already exists
			JobPath := config.GetSensitizePath(flag.Name)
			_, err = os.Stat(JobPath)
			if err != nil {
				fmt.Printf("%v sensitize.Delete failed, msg: Check name failed: Job [%v] is non-existent\n", ColorString("red", "[ERROR]"), flag.Name)
				os.Exit(1)
			}
			//Determine whether job can be deleted
			if file.IsJobRunning(sensitizeCsv, flag.Name) {
				fmt.Printf("%v Job %v is running, you can wait for it finishing or stop it.\n", ColorString("yellow", "[Warning]"), flag.Name)
				return
			}

			RunDeleteRemote(cmd.Context(), flag)
			return
		},
	}

	cmd.Flags().StringVarP(&flag.Name, "job", "j", "", "available sensitivity identification data, query by \"keentune sensitize jobs\"")

	return cmd
}

func initSensitizeConf() error {
	err := config.InitBrainConf()
	if err != nil {
		return err
	}

	log.Init()
	return nil
}

