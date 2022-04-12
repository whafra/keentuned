package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	com "keentune/daemon/api/common"
	"keentune/daemon/common/config"
	"keentune/daemon/common/log"
	"os"
	"strings"
)

const (
	egInfo         = "\tkeentune profile info --name cpu_high_load.conf"
	egSet          = "\tkeentune profile set --group1 cpu_high_load.conf"
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
	const GroupNum int = 20
	cmd := &cobra.Command{
		Use:     "set",
		Short:   "Apply a profile to the target machine",
		Long:    "Apply a profile to the target machine",
		Example: egSet,
		Run: func(cmd *cobra.Command, args []string) {
			//判断若args有值且以.conf结尾，则认为是默认所有group下发统一配置
			if len(args) > 0 && strings.HasSuffix(args[0], ".conf") {
				for i, _ := range setFlag.ConfFile {
					setFlag.Group[i] = true
					setFlag.ConfFile[i] = args[0]
				}
			} else {
				//若groupX已配置且以.conf结尾，则认为该配置有效
				for i, v := range setFlag.ConfFile {
					if len(v) != 0 && strings.HasSuffix(v, ".conf") {
						setFlag.Group[i] = true
					} else {
						setFlag.Group[i] = false
					}
				}
			}

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

	var group string = ""
	if err := initSet(); err != nil {
		setFlag.Group = make([]bool, GroupNum)
		setFlag.ConfFile = make([]string, GroupNum)
		for index := 0; index < GroupNum; index++ {
			group = fmt.Sprintf("group%d", index)
			cmd.Flags().StringVar(&setFlag.ConfFile[index], group, "", "profile name, query by command \"keentune profile list\"")
		}
	} else {
		setFlag.Group = make([]bool, len(config.KeenTune.Target.Group))
		setFlag.ConfFile = make([]string, len(config.KeenTune.Target.Group))
		for index, _ := range config.KeenTune.Target.Group {
			group = fmt.Sprintf("group%d", index+1)
			cmd.Flags().StringVar(&setFlag.ConfFile[index], group, "", "profile name, query by command \"keentune profile list\"")
		}
	}

	return cmd
}

func initSet() error {
	if err := config.InitTargetGroup(); err != nil {
		return err
	}

	log.Init()
	return nil
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

			err := config.InitWorkDir()
			if err != nil {
				fmt.Printf("%v Init work directory error: %v .\n", ColorString("red", "[ERROR]"), err)
				os.Exit(1)
			}

			WorkPath := config.GetProfileWorkPath(flag.Name)
			_, err = os.Stat(WorkPath)
			if err == nil {
				fmt.Printf("%s %s '%s' ?Y(yes)/N(no)", ColorString("yellow", "[Warning]"), deleteTips, flag.Name)
				if !confirm() {
					fmt.Println("[-] Give Up Delete")
					return
				}
				RunDeleteRemote(cmd.Context(), flag)
			} else {
				HomePath := config.GetProfileHomePath(flag.Name)
				_, err = os.Stat(HomePath)
				if err == nil {
					fmt.Printf("%v profile.Delete failed, msg: Check name failed: %v is not supported to delete\n", ColorString("red", "[ERROR]"), HomePath)
				} else {
					fmt.Printf("%v profile.Delete failed, msg: Check name failed: File [%v] is non-existent\n", ColorString("red", "[ERROR]"), flag.Name)
					os.Exit(1)
				}
			}

			return
		},
	}

	cmd.Flags().StringVar(&flag.Name, "name", "", "profile name, query by command \"keentune profile list\"")

	return cmd
}

func generateCmd() *cobra.Command {
	var genFlag GenFlag
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

			err := config.InitWorkDir()
			if err != nil {
				fmt.Printf("%s %v", ColorString("red", "[ERROR]"), err)
				os.Exit(1)
			}

			workPathName := config.GetProfileWorkPath(genFlag.Name)
			homePathName := config.GetProfileHomePath(genFlag.Name)
			_, err = ioutil.ReadFile(workPathName)
			if err != nil {
				_, errinfo := ioutil.ReadFile(homePathName)
				if errinfo != nil {
					fmt.Printf("%s profile.Generate failed, msg: Convert file: %v, read file :%v err:%v\n", ColorString("red", "[ERROR]"), genFlag.Name, homePathName, errinfo)
					os.Exit(1)
				}
			}

			//Determine whether json file already exists
			ParamPath := config.GetGenerateWorkPath(genFlag.Output)
			_, err = os.Stat(ParamPath)
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
