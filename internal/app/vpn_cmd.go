package app

import (
	"log"

	"github.com/spf13/cobra"

	"capstone_network_test/internal/diagnosis/vpn"
	"capstone_network_test/internal/mq"
)

var vpnCmd = &cobra.Command{
	Use:   "vpn",
	Short: "VPN 터널 · 경로 · 성능 · 보안 정책 점검",
	RunE:  runVPN,
}

func init() {
	rootCmd.AddCommand(vpnCmd)
}

func runVPN(cmd *cobra.Command, args []string) error {
	pub, err := mq.NewPublisher()
	if err != nil {
		log.Printf("[vpn] MQ 연결 실패, 진단은 계속 진행됩니다: %v", err)
	}
	defer pub.Close()

	vpn.Run(pub)
	return nil
}
