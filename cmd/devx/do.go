package main

import (
	"devopzilla.com/guku/internal/client"
	"github.com/spf13/cobra"
)

var doCmd = &cobra.Command{
	Use:   "do",
	Short: "do DevX magic for the specified environment",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := client.Run(args[0], configDir, stackPath, buildersPath); err != nil {
			return err
		}
		return nil
	},
}