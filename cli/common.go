package main

import (
	"fmt"
	"strings"

	"github.com/liushuochen/gotable"
	"github.com/spf13/cobra"
)

func subCommands() []*cobra.Command {
	var subCmds []*cobra.Command
	subCmds = append(subCmds, decorateCmd(createSensitizeCmds()))
	subCmds = append(subCmds, decorateCmd(createParamCmds()))
	subCmds = append(subCmds, decorateCmd(createProfileCmds()))
	subCmds = append(subCmds, decorateCmd(msgCmd()))

	return subCmds
}

func rollbackCmd(parentCmd string) *cobra.Command {
	var flag RollbackFlag
	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "rollback to the system init state",
		Long:  "\n\trollback the system config to the init state\n",
		RunE: func(cmd *cobra.Command, args []string) error {
			flag.Cmd = parentCmd
			return RunRollbackRemote(cmd.Context(), flag)
		},
	}

	return cmd
}

func deleteCmd(parentCmd string) *cobra.Command {
	var flag DeleteFlag
	cmd := &cobra.Command{
		Use:   "delete",
		Short: fmt.Sprintf("delete the specified %s info.", parentCmd),
		Long:  fmt.Sprintf("\n\tdelete the specified %s info.\n", parentCmd),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("\nrequired arguments: name=[%v]\n", flag.Name)
			if strings.Trim(flag.Name, " ") == "" {
				return fmt.Errorf("command flag --name is invalid argument")
			}

			flag.Cmd = parentCmd

			return RunDeleteRemote(cmd.Context(), flag)

		},
	}

	flags := cmd.Flags()
	flags.StringVar(&flag.Name, "name", "", fmt.Sprintf("the name of %s file you want to delete", parentCmd))

	return cmd
}

func setTuneFlag(cmdName string, cmd *cobra.Command, flag *TuneFlag) {
	flags := cmd.Flags()
	flags.StringVar(&flag.Name, "name", "", fmt.Sprintf("the %s name", cmdName))
	flags.IntVarP(&flag.Round, "iteration", "i", 100, fmt.Sprintf("the %s iterations", cmdName))
	flags.StringVar(&flag.BenchConf, "bench_conf", "", "the benchmark configure infomation")
	flags.StringVar(&flag.ParamConf, "param_conf", "parameter/sysctl.json", "the parameters infomation")
	flags.BoolVar(&flag.Verbose, "debug", false, "display all details, including baseline value, score set, average value, fluctuation range, execution time, tuning effect, etc.; by default, only the average value and tuning effect of target parameters with a weight of not 0 are displayed.")	
}

func msgCmd() *cobra.Command {
	var flag string
	var cmd = &cobra.Command{
		Use:   "msg",
		Short: "show the command executing result, enum: \"param tune\", \"sensitize collect\", \"sensitize train\"",
		Long:  "\n\tshow the command executing result, enum: \"param tune\", \"sensitize collect\", \"sensitize train\"\n\tfor example: keentune msg --name \"param tune\"\n",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("\nrequired arguments: name=[%v]\n", flag)
			if strings.Trim(flag, " ") == "" {
				return fmt.Errorf("command flag --name must be one of  \"param tune\", \"sensitize collect\", \"sensitize train\"")
			}
			
			return MsgRemote(cmd.Context(), flag)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&flag, "name", "", "the command name, enum: \"param tune\", \"sensitize collect\", \"sensitize train\"")

	return cmd
}

func stopCmd(flag string) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "stop",
		Short: fmt.Sprintf("stop a %s task in progress", flag),
		Long:  fmt.Sprintf("\n\tstop a %s task in progress\n", flag),
		RunE: func(cmd *cobra.Command, args []string) error {
			return StopRemote(cmd.Context(), flag)
		},
	}

	return cmd
}

// showInTable dispaly headers and params in table, params[n] length should be equal to headers'.
func showInTable(reply string) {
	headers, params := getHeaderAndParam(strings.Trim(reply, "\n"))
	tb, err := gotable.Create(headers...)
	if err != nil {
		fmt.Printf("create table failed: %v", err.Error())
		return
	}

	for _, rows := range params {
		if len(headers) != len(rows) {
			continue
		}

		tabeValue := make(map[string]string)
		for index, row := range rows {
			tabeValue[headers[index]] = row
		}

		err = tb.AddRow(tabeValue)
		if err != nil {
			fmt.Printf("add value to table failed: %v", err.Error())
			continue
		}
	}

	// align left
	for _, header := range headers {
		tb.Align(header, gotable.Left)
	}

	tb.PrintTable()
}

func getHeaderAndParam(reply string) ([]string, [][]string) {
	var headers []string
	var params [][]string
	complex := strings.Split(reply, ";")
	for index, info := range complex {
		if index == 0 {
			headers = append(headers, strings.Split(info, ",")...)
		} else {
			params = append(params, strings.Split(info, ","))
		}
	}

	return headers, params
}

// confirm Interactive reply on terminal: [true] same as yes; false same as no.
func confirm() bool {
	var inputInfo string
	for {
		fmt.Scanln(&inputInfo)
		if inputInfo != "y" && inputInfo != "n" && inputInfo != "yes" && inputInfo != "no" {
			fmt.Printf("\tplease input y(yes) or n(no)--> ")
			continue
		}

		if inputInfo == "y" || inputInfo == "yes" {
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

