package main

import (
	"fmt"
	"io/ioutil"
	"keentune/daemon/common/config"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const (
	exampleInfo         = "\tkeentune profile info --name cpu_high_load.conf"
	exampleSet  = "\tkeentune profile set --group1 cpu_high_load.conf\n" +
		"\tkeentune profile set cpu_high_load.conf"
	exampleGenerate     = "\tkeentune profile generate --name tune_test.conf --output gen_param_test.json"
	exampleProfDelete   = "\tkeentune profile delete --name tune_test.conf"
	exampleProfList     = "\tkeentune profile list"
	exampleProfRollback = "\tkeentune profile rollback"
)

// keentune profile
func createProfileCmds() *cobra.Command {
	var profCmd = &cobra.Command{
		Use:   "profile [command]",
		Short: "Static tuning with expert profiles",
		Long:  "Static tuning with expert profiles",
		Example: fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
			exampleProfDelete,
			exampleGenerate,
			exampleInfo,
			exampleProfList,
			exampleProfRollback,
			exampleSet,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Invaild argument. See help information: \n\n")
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

// keentune profile info
func infoCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:     "info",
		Short:   "Show information of the specified profile",
		Long:    "Show information of the specified profile",
		Example: exampleInfo,
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Trim(name, " ") == "" {
				fmt.Printf("Invaild argument. See help information: \n\n")
				cmd.Help()
			} else {
				name = strings.TrimSuffix(name, ".conf") + ".conf"
				RunInfoRemote(cmd.Context(), name)
			}
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "profile name, query by command \"keentune profile list\"")
	return cmd
}

// keentune profile list
func listProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all profiles",
		Long:    "List all profiles",
		Example: exampleProfList,
		Run: func(cmd *cobra.Command, args []string) {
			RunListRemote(cmd.Context(), "profile")
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
		Example: exampleSet,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 && setWithoutAnyGroup(setFlag.ConfFile) {
				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
				cmd.Help()
				return
			}

			// bind configuration file to group
			bindFileToGroup(args, setFlag)

			RunSetRemote(cmd.Context(), setFlag)
			return
		},
	}

	var group string
	if err := config.InitTargetGroup(); err != nil {
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

func setWithoutAnyGroup(groupFiles []string) bool {
	for _, fileName := range groupFiles {
		if len(fileName) != 0 {
			return false
		}
	}

	return true
}

func bindFileToGroup(args []string, setFlag SetFlag) {
	// Case1: bind all groups to the same configuration, when args passed. 
	if len(args) > 0 {
		for i, _ := range setFlag.ConfFile {
			setFlag.Group[i] = true
			setFlag.ConfFile[i] = args[0]
		}

		return
	}

	// Case2: bind a group according to the corresponding configuration by '--groupx' flag.
	for i, v := range setFlag.ConfFile {
		if strings.HasSuffix(v, ".conf") {
			setFlag.Group[i] = true
			continue
		}

		if len(v) != 0 {
			fmt.Printf("%v group%v, file %v is not  with .conf suffix.\n", ColorString("red", "[ERROR]"), i, v)
			os.Exit(1)
		}
	}

	return
}

func deleteProfileCmd() *cobra.Command {
	var flag DeleteFlag
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete a profile",
		Long:    "Delete a profile",
		Example: exampleProfDelete,
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Trim(flag.Name, " ") == "" {
				fmt.Printf("Invaild argument. See help information: \n\n")
				cmd.Help()
				return
			}

			flag.Cmd = "profile"
			flag.Name = strings.TrimSuffix(flag.Name, ".conf") + ".conf"

			initWorkDirectory()
			WorkPath := config.GetProfileWorkPath(flag.Name)
			_, err := os.Stat(WorkPath)
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
				}
				os.Exit(1)
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
		Example: exampleGenerate,
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Trim(genFlag.Name, " ") == "" {
				fmt.Printf("Invaild argument. See help information: \n\n")
				cmd.Help()
				return
			}

			genFlag.Name = strings.TrimSuffix(genFlag.Name, ".conf") + ".conf"
			if strings.Trim(genFlag.Output, " ") == "" {
				genFlag.Output = strings.TrimSuffix(genFlag.Name, ".conf") + ".json"
			} else {
				genFlag.Output = strings.TrimSuffix(genFlag.Output, ".json") + ".json"
			}

			initWorkDirectory()
			workPathName := config.GetProfileWorkPath(genFlag.Name)
			homePathName := config.GetProfileHomePath(genFlag.Name)
			_, err := ioutil.ReadFile(workPathName)
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

