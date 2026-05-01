package app

import (
	"log"

	"github.com/spf13/cobra"

	osdiag "capstone_network_test/internal/diagnosis/os"
	"capstone_network_test/internal/mq"
)

var osCmd = &cobra.Command{
	Use:   "os",
	Short: "OS 시스템 자원 점검 (CPU / Memory / Disk / Process)",
	RunE:  runOS,
}

func init() {
	rootCmd.AddCommand(osCmd)
}

func runOS(cmd *cobra.Command, args []string) error {
	pub, err := mq.NewPublisher()
	if err != nil {
		log.Printf("[os] MQ 연결 실패, 진단은 계속 진행됩니다: %v", err)
	}
	defer pub.Close()

	osdiag.Run(pub)
	return nil
}
