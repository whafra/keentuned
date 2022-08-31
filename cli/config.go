package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"keentune/daemon/common/config"
	"keentune/daemon/common/file"
	"os"
	"strings"
)

func createConfigCmd() *cobra.Command {
	var configCmd = &cobra.Command{
		Use:     "config [command]",
		Short:   "KeenTuned configuration operation command",
		Long:    "KeenTuned configuration operation command",
		Example: "keentune config target",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
			cmd.Usage()
			os.Exit(1)
		},
	}

	var availableCmd []*cobra.Command

	availableCmd = append(availableCmd, decorateCmd(targetCmd()))
	configCmd.AddCommand(availableCmd...)

	return configCmd
}

func targetCmd() *cobra.Command {
	var grep string
	cmd := &cobra.Command{
		Use:     "target",
		Short:   "Show target group configuration",
		Long:    "Show target group configuration",
		Example: "keentune config target",
		Run: func(cmd *cobra.Command, args []string) {
			RunTargetRemote(cmd.Context(), grep)
			return
		},
	}

	cmd.Flags().StringVar(&grep, "grep", "", "filter group with specified configuration")

	return cmd
}

func RunTargetRemote(context context.Context, grep string) {
	err := config.InitTargetGroup()
	if err != nil {
		fmt.Printf("%v cat target config: %v", ColorString("red", "[ERROR]"), err)
		os.Exit(1)
	}

	var reply string
	if grep == "" {
		for _, group := range config.KeenTune.Target.Group {
			reply += fmt.Sprintf("[target-group-%v]\n", group.GroupNo)
			reply += fmt.Sprintf("TARGET_IP = %v\n", strings.Join(group.IPs, ","))
		}

		fmt.Printf("%v", reply)
		return
	}

	filePath := config.GetProfileWorkPath("active.conf")
	activeGroup := file.GetRecord(filePath, "name", grep, "group_info")
	if len(activeGroup) == 0 {
		fmt.Println("No matching group found")
		return
	}

	actives := strings.Split(activeGroup, " ")
	for _, group := range config.KeenTune.Target.Group {
		for _, info := range actives {
			if strings.Contains(group.GroupName, strings.Trim(info, "group")) {
				reply += fmt.Sprintf("[target-group-%v]\n", group.GroupNo)
				reply += fmt.Sprintf("TARGET_IP = %v\n", strings.Join(group.IPs, ","))
				break
			}
		}
	}

	if reply == "" {
		fmt.Println("No matching group found")
		return
	}

	fmt.Printf("%v", reply)
	return
}

