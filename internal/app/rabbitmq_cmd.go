package app

import (
	"log"

	"github.com/spf13/cobra"

	"capstone_network_test/internal/diagnosis/rabbitmq"
	"capstone_network_test/internal/mq"
)

var rabbitmqCmd = &cobra.Command{
	Use:   "rabbitmq",
	Short: "RabbitMQ 큐 상태 점검",
	RunE:  runRabbitmq,
}

func init() {
	rootCmd.AddCommand(rabbitmqCmd)
}

func runRabbitmq(cmd *cobra.Command, args []string) error {
	pub, err := mq.NewPublisher()
	if err != nil {
		log.Printf("[rabbitmq] MQ 연결 실패, 진단은 계속 진행됩니다: %v", err)
	}
	defer pub.Close()

	rabbitmq.Run(pub)
	return nil
}
