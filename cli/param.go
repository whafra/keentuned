package main

import (
	"fmt"
	"strings"
	"github.com/spf13/cobra"
)

func createParamCmds() *cobra.Command {
	var paramCmd = &cobra.Command{
		Use:   "param",
		Short: "dynamic tuning commands, see details with '-h'",
		Long: "\n\tdynamic tuning commands, see details with '-h'",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	var paramCommands []*cobra.Command
	paramCommands = append(paramCommands, decorateCmd(tuneCmd()))
	paramCommands = append(paramCommands, decorateCmd(dumpCmd()))
	paramCommands = append(paramCommands, decorateCmd(listParamCmd()))
	paramCommands = append(paramCommands, decorateCmd(rollbackCmd("param")))
	paramCommands = append(paramCommands, decorateCmd(deleteCmd("param")))
	paramCommands = append(paramCommands, decorateCmd(stopCmd("param")))
	paramCmd.AddCommand(paramCommands...)

	return paramCmd
}

func tuneCmd() *cobra.Command {
	var flag TuneFlag
	cmd := &cobra.Command{
		Use:   "tune",
		Short: "dynamic sepecified algorithm search optimal parameter sets",
		Long: `
		tuning command using sepecified algorithm method dynamic search optimal parameter sets,
		the PROJECT_JSON which you can refer to Documentation xxx.json.json.
			example: keentune param tune --param_conf  parameter/sysctl.json --bench_conf benchmark/wrk/bench_wrk_nginx_long.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("\nrequired arguments: name=[%v], bench_conf=[%v]\n", flag.Name, flag.BenchConf)
			fmt.Printf("optional arguments: iteration=[%v], param_conf=[%v], debug=[%v]\n", flag.Round, flag.ParamConf, flag.Verbose)

			if strings.Trim(flag.Name, " ") == ""  {
				return fmt.Errorf("command flag --name is invalid argument, please check")
			}

			if strings.Trim(flag.BenchConf, " ") == ""  {
				return fmt.Errorf("command flag --bench_conf is invalid argument, please check")
			}

			return RunTuneRemote(cmd.Context(), flag)
		},
	}

	setTuneFlag("tune", cmd, &flag)
	return cmd
}

func listParamCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list current support parameters' file names",
		Long:  "\n\tlist current support parameters' file names\n",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunListRemote(cmd.Context(), "param")
		},
	}

	return cmd
}

func dumpCmd() *cobra.Command {
	var dump DumpFlag
	cmd := &cobra.Command{
		Use:   "dump",
		Short: "dump the specified tuning parameters to the specified profile",
		Long:  "\n\tdump the specified tuning parameters to the specified profile\n",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("\nrequired arguments: name=[%v], output=[%v]\n", dump.Name, dump.Output)
			if strings.Trim(dump.Name, " ") == "" {
				return fmt.Errorf("command option --name (-n) is invalid argument")
			}

			if strings.Trim(dump.Output, " ") == "" {
				return fmt.Errorf("command option --output (-o) is invalid argument")
			}

			return RunDumpRemote(cmd.Context(), dump)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&dump.Name, "name", "n", "", "the tuned task name")
	flags.StringVarP(&dump.Output, "output", "o", "", "the dumped profile file name")
	return cmd
}
