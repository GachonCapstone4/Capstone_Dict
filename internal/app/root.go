package app

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "diag-tool",
	Short: "캡스톤 서버 진단 도구",
}

func Execute() error {
	return rootCmd.Execute()
}
