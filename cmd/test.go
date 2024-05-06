package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version string
)

func init() {
	rootCmd.AddCommand(testCmd)
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Allows us to test the helper",
	Long:  "Allows us to test the helper",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello from test!")
	},
}
