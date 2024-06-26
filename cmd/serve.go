package cmd

import (
	"github.com/spf13/cobra"
	percontainerroles "github.com/thepatrick/per-docker-roles/per_container_roles"
)

var (
	port          int
	listenAddress string
	dockerNetwork string
)

func init() {
	rootCmd.AddCommand(serveCmd)
	// initCredentialsSubCommand(serveCmd)
	serveCmd.PersistentFlags().IntVar(&port, "port", percontainerroles.DefaultPort, "The port used to run the local server")
	serveCmd.PersistentFlags().StringVar(&listenAddress, "listen-on", percontainerroles.DefaultLocalHostAddress, "The address to listen on for incoming connections")
	serveCmd.PersistentFlags().StringVar(&dockerNetwork, "docker-network", percontainerroles.DefaultDockerNetwork, "The Docker network to use for container discovery")
}

var serveCmd = &cobra.Command{
	Use:   "serve [flags]",
	Short: "Serve AWS credentials through a local endpoint",
	Long:  "Serve AWS credentials through a local endpoint that is compatible with IMDSv2",
	Run: func(cmd *cobra.Command, args []string) {
		// helper.Debug = credentialsOptions.Debug

		percontainerroles.Serve(port, listenAddress, dockerNetwork) //, credentialsOptions
	},
}
