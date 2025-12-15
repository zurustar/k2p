package cli

import (
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kindle-to-pdf",
	Short: "Convert Kindle books to PDF format",
	Long: `A command-line tool that converts Kindle books (AZW, AZW3, MOBI formats) 
to PDF format on macOS systems using Calibre as the conversion backend.

The tool provides a simple interface for users to convert their personal 
Kindle library books while respecting DRM limitations and focusing on 
DRM-free content.`,
	Version: version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
	
	// Add version template
	rootCmd.SetVersionTemplate(`{{printf "%s version %s\n" .Name .Version}}`)
}