/*
Copyright © 2021 KeenTune

Package main for cli
*/
package main

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "keentune [command]",
	Short:   "KeenTune is an AI tuning tool for Linux system and cloud applications",
	Long:    "KeenTune is an AI tuning tool for Linux system and cloud applications",
	Example: "\tkeentune param -h\n\tkeentune profile -h\n\tkeentune sensitize -h",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

const template = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}{{end}}

Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}{{end}} {{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand)}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

func main() {
	decorateCmd(rootCmd)
	rootCmd.SetHelpTemplate(template)
	rootCmd.AddCommand(subCommands()...)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
