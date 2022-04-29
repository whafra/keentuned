package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"keentune/daemon/common/config"
	"os"
	"strings"
)

func createConfigCmd() *cobra.Command {
	var configCmd = &cobra.Command{
		Use:     "config [command]",
		Short:   "KeenTuned configuration operation command",
		Long:    "KeenTuned configuration operation command",
		Example: "keentune config target",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				if args[0] != "--help" && args[0] != "-h" && args[0] != "target" {
					fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
				}
			}

			if len(args) == 0 {
				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
			}

			return cmd.Help()
		},
	}

	var availableCmd []*cobra.Command

	availableCmd = append(availableCmd, decorateCmd(targetCmd()))
	configCmd.AddCommand(availableCmd...)

	return configCmd
}

func targetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "target",
		Short:   "Show target group configuration",
		Long:    "Show target group configuration",
		Example: "keentune config target",
		Run: func(cmd *cobra.Command, args []string) {
			RunTargetRemote(cmd.Context())
			return
		},
	}

	return cmd
}

func RunTargetRemote(context context.Context) {
	err := config.InitTargetGroup()
	if err != nil {
		fmt.Printf("%v cat target config: %v", ColorString("red", "[ERROR]"), err)
		os.Exit(1)
	}

	var reply string
	for _, group := range config.KeenTune.Target.Group {
		reply += fmt.Sprintf("[target-group-%v]\n", group.GroupNo)
		reply += fmt.Sprintf("TARGET_IP = %v\n", strings.Join(group.IPs, ","))
	}

	fmt.Printf("%v", reply)
	return
}
