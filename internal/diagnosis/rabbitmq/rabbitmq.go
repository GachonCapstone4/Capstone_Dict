package rabbitmq

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"capstone_network_test/internal/models"
	"capstone_network_test/internal/mq"
)

const (
	moduleName      = "rabbitmq"
	defaultHost     = "rabbitmq-headless.rabbitmq"
	defaultMgmtPort = "15672"
)

type queueInfo struct {
	Name                   string  `json:"name"`
	State                  string  `json:"state"`
	Consumers              int     `json:"consumers"`
	Messages               int     `json:"messages"`
	MessagesReady          int     `json:"messages_ready"`
	MessagesUnacknowledged int     `json:"messages_unacknowledged"`
	Durable                bool    `json:"durable"`
	AutoDelete             bool    `json:"auto_delete"`
	ConsumerUtilisation    float64 `json:"consumer_utilisation"`
}

func Run(pub mq.Publisher) {
	emit(pub, "start", models.StatusInfo, "RabbitMQ 큐 상태 점검을 시작합니다...", nil, 32)

	rawJSON, err := fetchQueues()
	if err != nil {
		emit(pub, "queues_result", models.StatusError,
			fmt.Sprintf("/api/queues 조회 실패: %v", err), nil, 68)
		emit(pub, "complete", models.StatusInfo, "RabbitMQ 큐 상태 점검이 완료되었습니다.", nil, 32)
		return
	}

	formatted := formatQueues(rawJSON)
	emit(pub, "queues_result", models.StatusOK, "모든 큐 상세 정보 조회 완료",
		map[string]string{"raw_output": formatted}, 68)
	emit(pub, "complete", models.StatusInfo, "RabbitMQ 큐 상태 점검이 완료되었습니다.", nil, 32)
}

func formatQueues(rawJSON string) string {
	var queues []queueInfo
	if err := json.Unmarshal([]byte(rawJSON), &queues); err != nil {
		return rawJSON
	}

	var sb strings.Builder
	for i, q := range queues {
		if i > 0 {
			sb.WriteString("\n")
		}
		fmt.Fprintf(&sb, "Queue: %s\n", q.Name)
		fmt.Fprintf(&sb, "  State:       %s\n", q.State)
		fmt.Fprintf(&sb, "  Consumers:   %d\n", q.Consumers)
		fmt.Fprintf(&sb, "  Messages:    %d\n", q.Messages)
		fmt.Fprintf(&sb, "  Ready:       %d\n", q.MessagesReady)
		fmt.Fprintf(&sb, "  Unacked:     %d\n", q.MessagesUnacknowledged)
		fmt.Fprintf(&sb, "  Durable:     %v\n", q.Durable)
		fmt.Fprintf(&sb, "  Auto-delete: %v", q.AutoDelete)
	}
	return sb.String()
}

func fetchQueues() (string, error) {
	user := os.Getenv("RABBITMQ_USER")
	pass := os.Getenv("RABBITMQ_PASS")
	host := os.Getenv("RABBITMQ_MGMT_HOST")
	port := os.Getenv("RABBITMQ_MGMT_PORT")

	if host == "" {
		host = defaultHost
	}
	if port == "" {
		port = defaultMgmtPort
	}

	url := fmt.Sprintf("http://%s:%s/api/queues", host, port)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("요청 생성 실패: %w", err)
	}
	req.SetBasicAuth(user, pass)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d 응답", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("응답 읽기 실패: %w", err)
	}

	return string(body), nil
}

func emit(pub mq.Publisher, stage, status, message string, data interface{}, bannerWidth int) {
	printBanner(message, bannerWidth)

	msg := models.DiagMessage{
		UserID:    1,
		SSEType:   "rabbitmq",
		Module:    moduleName,
		Stage:     stage,
		Status:    status,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().UTC(),
	}

	if err := pub.Publish(msg); err != nil {
		log.Printf("[MQ] publish 실패 (stage=%s): %v", stage, err)
	}
}

func printBanner(message string, width int) {
	sep := strings.Repeat("=", width)
	fmt.Println(sep)
	fmt.Println(message)
	fmt.Println(sep)
}
