package main

import (
	"fmt"
	"strings"
	"github.com/spf13/cobra"
)

func createProfileCmds() *cobra.Command {
	var profCmd = &cobra.Command{
		Use:   "profile",
		Short: "static tuning commands, see details with '-h'",
		Long: "\n\tstatic tuning commands, see details with '-h'",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	var profileCommands []*cobra.Command
	profileCommands = append(profileCommands, decorateCmd(infoCmd()))
	profileCommands = append(profileCommands, decorateCmd(setCmd()))
	profileCommands = append(profileCommands, decorateCmd(deleteCmd("profile")))
	profileCommands = append(profileCommands, decorateCmd(listProfileCmd()))
	profileCommands = append(profileCommands, decorateCmd(rollbackCmd("profile")))
	profileCommands = append(profileCommands, decorateCmd(generateCmd()))

	profCmd.AddCommand(profileCommands...)
	return profCmd
}

func infoCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "info",
		Short: "cat information of specified profile",
		Long:  "\n\tcat information of specified profile\n\texample: keentune param info --name cpu_high_load.conf\n",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("\nrequired arguments: name=[%v]\n", name)
			if strings.Trim(name, " ") == "" {
				return fmt.Errorf("command option --name is invalid argument")
			}

			return RunInfoRemote(cmd.Context(), name)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&name, "name", "", "the name of the browse profile file")

	return cmd
}

func listProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all the supported profile",
		Long:  "\n\tlist all the supported profile\n",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunListRemote(cmd.Context(), "profile")
		},
	}

	return cmd
}

func setCmd() *cobra.Command {
	var setFlag SetFlag
	cmd := &cobra.Command{
		Use:   "set",
		Short: "apply the specified profile to target machine",
		Long: "\n\tapply the specified profile to target machine\n\texample: keentune profile set --name cpu_high_load.conf\n",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("\nrequired arguments: name=[%v]\n", setFlag.Name)
			
			if strings.Trim(setFlag.Name, " ") == "" {
				return fmt.Errorf("command option --name is invalid argument")
			}

			return RunSetRemote(cmd.Context(), setFlag)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&setFlag.Name, "name", "", "the specified profile name")

	return cmd
}

func generateCmd() *cobra.Command {
	var genFlag DumpFlag
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "generate profile file from specified parameters json file",
		Long:  "\n\tgenerate profile file from specified parameters json file\n",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("\nrequired arguments: name=[%v], output=[%v]\n", genFlag.Name, genFlag.Output)
			if strings.Trim(genFlag.Name, " ") == "" {
				return fmt.Errorf("command option --name (-n) is invalid argument")
			}

			if strings.Trim(genFlag.Output, " ") == "" {
				return fmt.Errorf("command option --output (-o) is invalid argument")
			}
			return RunGenerateRemote(cmd.Context(), genFlag)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&genFlag.Name, "name", "n", "", "the profile set file name")
	flags.StringVarP(&genFlag.Output, "output", "o", "", "the dumping json file name")

	return cmd
}
