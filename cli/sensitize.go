package main

import (
	"fmt"
	"strings"
	"github.com/spf13/cobra"
)

func createSensitizeCmds() *cobra.Command {
	sensitizeCmd := &cobra.Command{
		Use:   "sensitize",
		Short: "sensitive paramater recognization commands, see details with '-h'",
		Long: "\n\tsensitive paramater recognization commands, see details with '-h'",
		RunE: func(cmd *cobra.Command, args []string) error {
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
		Use:   "collect",
		Short: "perform sensitive parameter collection",
		Long: `
		perform sensitive parameter collection, the PROJECT_JSON which you can refer to Documentation xxx.json.json.
			example: keentune sensitize --name g6_long_latency --param_conf  parameter/sysctl.json --bench_conf benchmark/wrk/bench_wrk_nginx_long.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("\nrequired arguments: name=[%v], bench_conf=[%v]\n", flag.Name, flag.BenchConf)
			fmt.Printf("optional arguments: iteration=[%v], param_conf=[%v], debug=[%v]\n", flag.Round, flag.ParamConf, flag.Verbose)

			if strings.Trim(flag.Name, " ") == "" {
				return fmt.Errorf("command flag --name is invalid argument")
			}

			if strings.Trim(flag.BenchConf, " ") == "" {
				return fmt.Errorf("command flag --bench_conf is invalid argument")
			}

			return RunCollectRemote(cmd.Context(), flag)
		},
	}

	setTuneFlag("sensitize", cmd, &flag)

	return cmd
}

func trainCmd() *cobra.Command {
	var trainflags TrainFlag
	cmd := &cobra.Command{
		Use:   "train",
		Short: "train sensitive parameters",
		Long:  "\n\ttrain sensitive parameters\n",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("\nrequired arguments: data_name=[%v], output_name=[%v]\n", trainflags.Data, trainflags.Output)
			fmt.Printf("optional arguments: trials=[%v]\n", trainflags.Trials)

			if strings.Trim(trainflags.Data, " ") == "" {
				return fmt.Errorf("command option --data_name is invalid argument")
			}

			if strings.Trim(trainflags.Output, " ") == "" {
				return fmt.Errorf("command option --output_name is invalid argument")
			}

			if trainflags.Trials > 10 || trainflags.Trials < 1 {
				return fmt.Errorf("command option --trials is out of range [1,10], pleast check and try again")
			}

			return RunTrainRemote(cmd.Context(), trainflags)
		},
	}

	flags := cmd.Flags()
	flags.IntVarP(&trainflags.Trials, "trials", "t", 1, "the sensitize trials")
	flags.StringVar(&trainflags.Data, "data_name", "", "the specified training sensitive data")
	flags.StringVar(&trainflags.Output, "output_name", "", "the training result output name")

	return cmd
}

func listSensitivityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list brief info of sensitive data",
		Long:  "\n\tlist brief info of sensitive data\n",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunListRemote(cmd.Context(), "sensitize")
		},
	}

	return cmd
}

func deleteSensitivityCmd() *cobra.Command {
	var flag DeleteFlag
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete the specified sensitive parameters",
		Long:  "\n\tdelete the specified sensitive parameters\n",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("\nrequired arguments: name=[%v]\n", flag.Name)
			if strings.Trim(flag.Name, " ") == "" {
				return fmt.Errorf("command option --name is invalid argument")
			}			
			
			flag.Cmd = "sensitize"
			return RunDeleteRemote(cmd.Context(), flag)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&flag.Name, "name", "", "the sensitize data wanted to delete")

	return cmd
}
