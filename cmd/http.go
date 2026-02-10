package cmd

import (
	"github.com/spf13/cobra"
)

var ociPath string

var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "HTTP Server for Skills Registry",
	Long:  `Commands for serving and generating the Skills Registry interface via HTTP.`,
}

func init() {
	rootCmd.AddCommand(httpCmd)
	httpCmd.PersistentFlags().StringVar(&ociPath, "oci-path", "", "Path to the OCI store (default: system store)")
}
