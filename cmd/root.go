package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "per-docker-roles [command]",
	Short: "A proxy to aws_signing_helper, providing a role based on the calling docker container",
	Long: `A tool that looks up the role of the calling docker container (based on calling IP and
label on that container), proxies to aws_signing_helper and then assumes the requested role.`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
