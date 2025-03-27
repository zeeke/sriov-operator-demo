/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zeeke/sriov-operator-demo/internal/scenarios"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available scenarios to generate",
	Run: func(cmd *cobra.Command, args []string) {
		for s := range scenarios.Index {
			fmt.Println(s)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
