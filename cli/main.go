/*
Copyright © 2021 KeenTune

Package main for cli
*/
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "keentune",
	Short: "\n\tkeentune is a command line tool for AI tuning system.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// rootCmd help command
var helpCmd = &cobra.Command{
	Use:   "help",
	Short: "help message",
	Run: func(command *cobra.Command, args []string) {
		rootCmd.SetOutput(os.Stdout)
		_ = rootCmd.Usage()
	},
}

const keentuneLogo = " " +
"    __ __                  ______                  \n " +
"   / //_/___   ___   ____ /_  __/__  __ ____   ___ \n " +
"  / ,<  / _ \\ / _ \\ / __ \\ / /  / / / // __ \\ / _ \\ \n " +
" / /| |/  __//  __// / / // /  / /_/ // / / //  __/\n " +
"/_/ |_|\\___/ \\___//_/ /_//_/   \\__,_//_/ /_/ \\___/\n"

func main() {
	fmt.Print(keentuneLogo)
	decorateCmd(rootCmd)
	rootCmd.SetHelpCommand(helpCmd)
	rootCmd.AddCommand(subCommands()...)	

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
