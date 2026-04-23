package app

import (
	"log"

	"github.com/spf13/cobra"

	"capstone_network_test/internal/diagnosis/network"
	"capstone_network_test/internal/mq"
)

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "L2→L3→외부 물리 네트워크 점검",
	RunE:  runNetwork,
}

func init() {
	rootCmd.AddCommand(networkCmd)
}

func runNetwork(cmd *cobra.Command, args []string) error {
	pub, err := mq.NewPublisher()
	if err != nil {
		log.Printf("[network] MQ 연결 실패, 진단은 계속 진행됩니다: %v", err)
	}
	defer pub.Close()

	network.Run(pub)
	return nil
}
