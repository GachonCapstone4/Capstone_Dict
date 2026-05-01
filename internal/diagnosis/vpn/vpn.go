package vpn

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"capstone_network_test/internal/models"
	"capstone_network_test/internal/mq"
)

const moduleName = "vpn"

func Run(pub mq.Publisher) {
	nodeIP := os.Getenv("NODE_IP")
	if nodeIP == "" {
		nodeIP = "unknown"
	}

	tracerouteTarget := os.Getenv("TRACEROUTE_TARGET")
	iperfTarget := os.Getenv("IPERF_TARGET")

	// Stage 1: Start
	emit(pub, nodeIP, "start", models.StatusInfo, "VPN 상태 점검을 시작합니다...", nil)

	// Stage 2: Path Validation (traceroute)
	emit(pub, nodeIP, "traceroute_start", models.StatusInfo,
		fmt.Sprintf("경로 검증을 시작합니다 (target: %s)...", tracerouteTarget), nil)
	traceResult, err := CheckTraceroute(tracerouteTarget)
	trStatus, trMessage := deriveTracerouteStatus(traceResult, err)
	emit(pub, nodeIP, "traceroute_result", trStatus, trMessage, traceResult)

	// Stage 3: Performance Quality (iperf3 -c [IP] -u)
	emit(pub, nodeIP, "iperf_start", models.StatusInfo,
		fmt.Sprintf("성능 품질을 측정합니다 (target: %s)...", iperfTarget), nil)
	iperfResult, err := CheckIperf(iperfTarget)
	iStatus, iMessage := deriveIperfStatus(iperfResult, err)
	emit(pub, nodeIP, "iperf_result", iStatus, iMessage, iperfResult)

	// Stage 4: Complete
	emit(pub, nodeIP, "complete", models.StatusInfo, "VPN 점검이 완료되었습니다.", nil)
}

func emit(pub mq.Publisher, nodeIP, stage, status, message string, data interface{}) {
	printBanner(message)

	msg := models.DiagMessage{
		UserID:    1,
		SSEType:   "vpn",
		Module:    moduleName,
		NodeIP:    nodeIP,
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

func printBanner(message string) {
	sep := strings.Repeat("=", 60)
	fmt.Println(sep)
	fmt.Println(message)
	fmt.Println(sep)
}

func deriveTracerouteStatus(r models.TracerouteResult, err error) (string, string) {
	if err != nil {
		return models.StatusError, fmt.Sprintf("traceroute 실행 오류: %v", err)
	}
	if r.RawOutput == "" {
		return models.StatusError, "traceroute 출력 없음"
	}
	return models.StatusOK, fmt.Sprintf("경로 검증 완료 (target: %s)", r.Target)
}

func deriveIperfStatus(r models.IperfResult, err error) (string, string) {
	if err != nil {
		return models.StatusError, fmt.Sprintf("iperf3 실행 오류: %v", err)
	}
	if r.TCPRawOutput == "" && r.UDPRawOutput == "" {
		return models.StatusError, "iperf3 출력 없음"
	}
	message := fmt.Sprintf("대역폭 측정 완료 (target: %s)\n\n[TCP -M 1380]\n%s\n\n[UDP -u -t 5]\n%s",
		r.Target, r.TCPRawOutput, r.UDPRawOutput)
	return models.StatusOK, message
}
