/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/zeeke/sriov-operator-demo/internal"
)

var scenario string

var rootCmd = &cobra.Command{
	Use:          "sriov-operator-demo",
	Short:        "Generate sample scenarios for the SR-IOV Network Operator",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if scenario == "" {
			return fmt.Errorf("--scenario parameter is mandatory")
		}

		f, ok := internal.Scenarios[scenario]
		if !ok {
			return fmt.Errorf("scenario [%s] not found", scenario)
		}

		return internal.DumpScenario(f)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	//flag.Set("v", "100")
	flag.Set("logtostderr", "true")
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	rootCmd.Flags().StringVarP(&scenario, "scenario", "s", "", "The scenario to generate")
}
