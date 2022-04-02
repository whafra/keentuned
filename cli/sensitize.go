package main

import (
	"fmt"
	"strings"
	"time"
	"os"
	"github.com/spf13/cobra"
	com "keentune/daemon/api/common"
)

const (
	egCollect       = "\tkeentune sensitize collect --param sysctl.json --bench bench_wrk_nginx_long.json --data collect_test --iteration 10"
	egTrain         = "\tkeentune sensitize train --data collect_test --output train_test --trials 2"
	egDelete        = "\tkeentune sensitize delete --data collect_test"
	egSensitiveList = "\tkeentune sensitize list"
	egSensitiveStop = "\tkeentune sensitize stop"
)

func createSensitizeCmds() *cobra.Command {
	sensitizeCmd := &cobra.Command{
		Use:     "sensitize [command]",
		Short:   "Sensitive parameter identification and explanation with AI algorithms",
		Long:    "Sensitive parameter identification and explanation with AI algorithms",
		Example: fmt.Sprintf("%s\n%s\n%s\n%s\n%s", egCollect, egDelete, egSensitiveList, egSensitiveStop, egTrain),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				if args[0] != "--help" && args[0] != "-h" && args[0] != "collect" && args[0] != "list" && args[0] != "delete" && args[0] != "train" && args[0] != "stop" {
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

	sesiCmds = append(sesiCmds, decorateCmd(collectCmd()))
	sesiCmds = append(sesiCmds, decorateCmd(listSensitivityCmd()))
	sesiCmds = append(sesiCmds, decorateCmd(trainCmd()))
	sesiCmds = append(sesiCmds, decorateCmd(deleteSensitivityCmd()))
	sesiCmds = append(sesiCmds, decorateCmd(stopCmd("sensitize")))

	sensitizeCmd.AddCommand(sesiCmds...)

	return sensitizeCmd
}

func collectCmd() *cobra.Command {
	var flag TuneFlag
	cmd := &cobra.Command{
		Use:     "collect",
		Short:   "Collecting parameter and benchmark score as sensitivity identification data randomly",
		Long:    "Collecting parameter and benchmark score as sensitivity identification data randomly",
		Example: egCollect,
		PreRun: func(cmd *cobra.Command, args []string) {
			if err := checkTuningFlags("sensitize", &flag); err != nil {
				fmt.Printf("%v check input: %v\n", ColorString("red", "[ERROR]"), err)
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Trim(flag.Name, " ") == "" || strings.Trim(flag.BenchConf, " ") == "" || len(flag.ParamConf) == 0 {
				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
				cmd.Help()
				return
			}

			flag.Log = fmt.Sprintf("%v/%v-%v.log", "/var/log/keentune", "keentuned-sensitize-collect", time.Now().Unix())
			RunCollectRemote(cmd.Context(), flag)
			return
		},
	}

	setTuneFlag("sensitize", cmd, &flag)
	return cmd
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

			if strings.Trim(trainflags.Output, " ") == "" {
				trainflags.Output = trainflags.Data
			}

			if trainflags.Trials > 10 || trainflags.Trials < 1 {
				fmt.Printf("%v Incomplete or Unmatched command, trials is out of range [1,10]\n\n", ColorString("red", "[ERROR]"))
				return
			}

			trainflags.Log = fmt.Sprintf("%v/%v-%v.log", "/var/log/keentune", "keentuned-sensitize-train", time.Now().Unix())

			RunTrainRemote(cmd.Context(), trainflags)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&trainflags.Data, "data", "d", "", "available sensitivity identification data, query by \"keentune sensitize list\"")
	flags.IntVarP(&trainflags.Trials, "trials", "t", 1, "sensitize trials")
	flags.StringVarP(&trainflags.Output, "output", "o", "", "output file of sensitive parameter identification and explanation")

	return cmd
}

func listSensitivityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List available sensitivity identification data",
		Long:    "List available sensitivity identification data",
		Example: egSensitiveList,
		Run: func(cmd *cobra.Command, args []string) {
			RunListRemote(cmd.Context(), "sensitize")
			return
		},
	}

	return cmd
}

func deleteSensitivityCmd() *cobra.Command {
	var flag DeleteFlag
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete the sensitivity identification data",
		Long:    "Delete the sensitivity identification data",
		Example: egDelete,
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Trim(flag.Name, " ") == "" {
				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
				cmd.Help()
				return
			}

			_, _, DataList, err := com.GetDataList()
                        if err != nil {
                                if find := strings.Contains(err.Error(), "connection refused"); find {
                                        fmt.Println("brain access denied")
                                        return
                                }
                                fmt.Println("Get sensitize Data List err:%v", err)
                                return
                        }
                        if find := strings.Contains(DataList, flag.Name); find {
                                fmt.Printf("%s %s '%s' ?Y(yes)/N(no)", ColorString("yellow", "[Warning]"), deleteTips, flag.Name)
                                if !confirm() {
                                        fmt.Println("[-] Give Up Delete")
                                        return
                                }
                                flag.Cmd = "sensitize"
                                RunDeleteRemote(cmd.Context(), flag)
                        } else {
				err := fmt.Sprintf("Sensitize delete failed: File %s is non-existent", flag.Name)
				fmt.Printf("%s %s\n", ColorString("red", "[ERROR]"), err)
                        }

                        return
		},
	}

	cmd.Flags().StringVarP(&flag.Name, "data", "d", "", "available sensitivity identification data, query by \"keentune sensitize list\"")

	return cmd
}
