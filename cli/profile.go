package main

import (
	"fmt"
	"github.com/spf13/cobra"
	com "keentune/daemon/api/common"
	m "keentune/daemon/modules"
	"os"
	"strings"
)

const (
	egInfo         = "\tkeentune profile info --name cpu_high_load.conf"
	egSet          = "\tkeentune profile set --name cpu_high_load.conf"
	egGenerate     = "\tkeentune profile generate --name tune_test.conf --output gen_param_test.json"
	egProfDelete   = "\tkeentune profile delete --name tune_test.conf"
	egProfList     = "\tkeentune profile list"
	egProfRollback = "\tkeentune profile rollback"
)

func createProfileCmds() *cobra.Command {
	var profCmd = &cobra.Command{
		Use:     "profile [command]",
		Short:   "Static tuning with expert profiles",
		Long:    "Static tuning with expert profiles",
		Example: fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s", egProfDelete, egGenerate, egInfo, egProfList, egProfRollback, egSet),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				if args[0] != "--help" && args[0] != "-h" && args[0] != "generate" && args[0] != "list" && args[0] != "set" && args[0] != "delete" && args[0] != "info" && args[0] != "rollback" {
					fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
				}
			}

			if len(args) == 0 {
				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
			}

			return cmd.Help()
		},
	}

	var profileCommands []*cobra.Command
	profileCommands = append(profileCommands, decorateCmd(infoCmd()))
	profileCommands = append(profileCommands, decorateCmd(setCmd()))
	profileCommands = append(profileCommands, decorateCmd(deleteProfileCmd()))
	profileCommands = append(profileCommands, decorateCmd(listProfileCmd()))
	profileCommands = append(profileCommands, decorateCmd(rollbackCmd("profile")))
	profileCommands = append(profileCommands, decorateCmd(generateCmd()))

	profCmd.AddCommand(profileCommands...)
	return profCmd
}

func infoCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:     "info",
		Short:   "Show information of the specified profile",
		Long:    "Show information of the specified profile",
		Example: egInfo,
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Trim(name, " ") == "" {
				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
				cmd.Help()
				return
			}

			name = strings.TrimSuffix(name, ".conf") + ".conf"
			RunInfoRemote(cmd.Context(), name)
			return
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "profile name, query by command \"keentune profile list\"")

	return cmd
}

func listProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all profiles",
		Long:    "List all profiles",
		Example: egProfList,
		Run: func(cmd *cobra.Command, args []string) {
			RunListRemote(cmd.Context(), "profile")
			return
		},
	}

	return cmd
}

func setCmd() *cobra.Command {
	var setFlag SetFlag
	cmd := &cobra.Command{
		Use:     "set",
		Short:   "Apply a profile to the target machine",
		Long:    "Apply a profile to the target machine",
		Example: egSet,
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Trim(setFlag.Name, " ") == "" {
				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
				cmd.Help()
				return
			}

			setFlag.Name = strings.TrimSuffix(setFlag.Name, ".conf") + ".conf"
			var targetMsg = new(string)
			if com.IsTargetOffline(targetMsg) {
				fmt.Printf("%v Found %v offline, please get them (it) ready before use\n",
					ColorString("red", "[ERROR]"),
					strings.TrimSuffix(*targetMsg, ", "))
				os.Exit(1)
			}

			RunSetRemote(cmd.Context(), setFlag)
			return
		},
	}

	cmd.Flags().StringVar(&setFlag.Name, "name", "", "profile name, query by command \"keentune profile list\"")
	return cmd
}

func deleteProfileCmd() *cobra.Command {
	var flag DeleteFlag
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete a profile",
		Long:    "Delete a profile",
		Example: egProfDelete,
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Trim(flag.Name, " ") == "" {
				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
				cmd.Help()
				return
			}

			flag.Cmd = "profile"
			flag.Name = strings.TrimSuffix(flag.Name, ".conf") + ".conf"

			ProfilePath := m.GetProfileWorkPath(flag.Name)
			_, err := os.Stat(ProfilePath)
                        if err == nil {
                                fmt.Printf("%s %s '%s' ?Y(yes)/N(no)", ColorString("yellow", "[Warning]"), deleteTips, flag.Name)
                                if !confirm() {
                                        fmt.Println("[-] Give Up Delete")
                                        return
                                }
                                RunDeleteRemote(cmd.Context(), flag)
                        } else {
                                RunDeleteRemote(cmd.Context(), flag)
                        }

                        return
		},
	}

	cmd.Flags().StringVar(&flag.Name, "name", "", "profile name, query by command \"keentune profile list\"")

	return cmd
}

func generateCmd() *cobra.Command {
	var genFlag DumpFlag
	cmd := &cobra.Command{
		Use:     "generate",
		Short:   "Generate a parameter configuration file from profile",
		Long:    "Generate a parameter configuration file from profile",
		Example: egGenerate,
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Trim(genFlag.Name, " ") == "" {
				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
				cmd.Help()
				return
			}

			genFlag.Name = strings.TrimSuffix(genFlag.Name, ".conf") + ".conf"
			if strings.Trim(genFlag.Output, " ") == "" {
				genFlag.Output = strings.TrimSuffix(genFlag.Name, ".conf") + ".json"
			} else {
				genFlag.Output = strings.TrimSuffix(genFlag.Output, ".json") + ".json"
			}

			//Determine whether json file already exists
			ParamPath := m.GetGenerateWorkPath(genFlag.Output)
			_, err := os.Stat(ParamPath)
			if err == nil {
				fmt.Printf("%s %s", ColorString("yellow", "[Warning]"), fmt.Sprintf(outputTips, "generated parameter"))
				genFlag.Force = confirm()
				if !genFlag.Force {
					fmt.Printf("outputFile exist and you have given up to overwrite it\n")
					os.Exit(1)
				}
				RunGenerateRemote(cmd.Context(), genFlag)
			} else {
				RunGenerateRemote(cmd.Context(), genFlag)
			}

			return
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&genFlag.Name, "name", "n", "", "profile name, query by command \"keentune profile list\"")
	flags.StringVarP(&genFlag.Output, "output", "o", "", "output parameter configuration file name, default with suffix \".json\"")

	return cmd
}
