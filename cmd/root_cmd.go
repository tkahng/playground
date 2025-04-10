package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: filepath.Base(os.Args[0]),
}

func Execute() {
	rootCmd.Execute()
}
