package app

import (
	"log"

	"github.com/spf13/cobra"

	"capstone_network_test/internal/diagnosis/k8s"
	"capstone_network_test/internal/mq"
)

var k8sCmd = &cobra.Command{
	Use:   "k8s",
	Short: "Kubernetes 클러스터 상태 점검",
	RunE:  runK8s,
}

func init() {
	rootCmd.AddCommand(k8sCmd)
}

func runK8s(cmd *cobra.Command, args []string) error {
	pub, err := mq.NewPublisher()
	if err != nil {
		log.Printf("[k8s] MQ 연결 실패, 진단은 계속 진행됩니다: %v", err)
	}
	defer pub.Close()

	k8s.Run(pub)
	return nil
}
