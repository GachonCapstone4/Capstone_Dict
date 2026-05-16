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

func Run(pub mq.Publisher) {
	emit(pub, "start", models.StatusInfo, "RabbitMQ 큐 상태 점검을 시작합니다...", nil, 32)

	rawJSON, err := fetchQueues()
	if err != nil {
		emit(pub, "queues_result", models.StatusError,
			fmt.Sprintf("/api/queues 조회 실패: %v", err), nil, 68)
		emit(pub, "complete", models.StatusInfo, "RabbitMQ 큐 상태 점검이 완료되었습니다.", nil, 32)
		return
	}

	var queues []interface{}
	if jsonErr := json.Unmarshal([]byte(rawJSON), &queues); jsonErr != nil {
		emit(pub, "queues_result", models.StatusOK, "모든 큐 상세 정보 조회 완료",
			map[string]string{"raw_output": rawJSON}, 68)
		emit(pub, "complete", models.StatusInfo, "RabbitMQ 큐 상태 점검이 완료되었습니다.", nil, 32)
		return
	}

	data := map[string]interface{}{"queues": queues}
	emit(pub, "queues_result", models.StatusOK, "모든 큐 상세 정보 조회 완료", data, 68)
	emit(pub, "complete", models.StatusInfo, "RabbitMQ 큐 상태 점검이 완료되었습니다.", nil, 32)
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
