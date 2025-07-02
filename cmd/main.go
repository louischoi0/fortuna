package cmd

import (
        "log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
        Use:   "ftn",
        Short: "ftn is a client for the MX trading engine",
}

func Execute() {
        if err := rootCmd.Execute(); err != nil {
                log.Fatalf("cli execution failed: %v", err)
        }
}

// Execute sets up the CLI entry point
func init() {
        nodeCMD := CreateNodeCMD()
        rootCmd.AddCommand(nodeCMD)
}
