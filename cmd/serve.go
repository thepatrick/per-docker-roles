package cmd

import (
	"github.com/spf13/cobra"
	percontainerroles "github.com/thepatrick/per-docker-roles/per_container_roles"
)

var (
	port int
)

func init() {
	rootCmd.AddCommand(serveCmd)
	// initCredentialsSubCommand(serveCmd)
	serveCmd.PersistentFlags().IntVar(&port, "port", percontainerroles.DefaultPort, "The port used to run the local server")
}

var serveCmd = &cobra.Command{
	Use:   "serve [flags]",
	Short: "Serve AWS credentials through a local endpoint",
	Long:  "Serve AWS credentials through a local endpoint that is compatible with IMDSv2",
	Run: func(cmd *cobra.Command, args []string) {
		// err := PopulateCredentialsOptions()
		// if err != nil {
		// 	log.Println(err)
		// 	os.Exit(1)
		// }

		// helper.Debug = credentialsOptions.Debug

		percontainerroles.Serve(port) //, credentialsOptions
	},
}
