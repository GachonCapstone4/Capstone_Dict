package network

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"capstone_network_test/internal/models"
	"capstone_network_test/internal/mq"
)

const (
	moduleName       = "network"
	defaultGatewayIP = "192.168.2.1"
	externalIP       = "8.8.8.8"
)

// buildRingMap reads node IPs from ConfigMap-injected env vars and constructs
// the ARP ring: controlplane → workernode-1 → workernode-2 → controlplane
func buildRingMap() (map[string]string, error) {
	cp := os.Getenv("controlplane")
	w1 := os.Getenv("workernode-1")
	w2 := os.Getenv("workernode-2")

	if cp == "" || w1 == "" || w2 == "" {
		return nil, fmt.Errorf("Ring 구성 환경변수 미설정 (controlplane=%q, workernode-1=%q, workernode-2=%q)", cp, w1, w2)
	}

	return map[string]string{
		cp: w1,
		w1: w2,
		w2: cp,
	}, nil
}

func Run(pub mq.Publisher) {
	nodeIP, err := detectNodeIP()
	if err != nil {
		log.Printf("[network] 노드 IP 감지 실패: %v", err)
		emit(pub, "unknown", "start", models.StatusError, fmt.Sprintf("노드 IP 감지 실패: %v", err), nil, 32)
		return
	}

	nodeNext, err := buildRingMap()
	if err != nil {
		log.Printf("[network] Ring 맵 구성 실패: %v", err)
		emit(pub, nodeIP, "start", models.StatusError, fmt.Sprintf("Ring 맵 구성 실패: %v", err), nil, 32)
		return
	}

	gatewayIP := os.Getenv("GATEWAY_IP")
	if gatewayIP == "" {
		gatewayIP = defaultGatewayIP
	}

	// Stage 1: Start
	emit(pub, nodeIP, "start", models.StatusInfo,
		fmt.Sprintf("%s의 물리 네트워크 점검을 시작합니다....", nodeIP), nil, 32)

	// Stage 2: ARP
	targetIP, ok := nodeNext[nodeIP]
	if !ok {
		emit(pub, nodeIP, "arp_result", models.StatusError,
			fmt.Sprintf("노드 IP %s는 알려진 Ring 구성원이 아닙니다", nodeIP), nil, 68)
	} else {
		emit(pub, nodeIP, "arp_start", models.StatusInfo,
			fmt.Sprintf("%s에서 %s Bridge 내부 ARP 테이블을 점검합니다...", nodeIP, targetIP), nil, 68)

		arpResult, err := CheckARP(targetIP)
		status, message := deriveARPStatus(arpResult, err)
		emit(pub, nodeIP, "arp_result", status, message, arpResult, 68)
	}

	// Stage 3: Gateway ICMP
	gwResult, err := CheckPing(gatewayIP)
	gwStatus, gwMessage := derivePingStatus(gwResult, err)
	gwFullMsg := fmt.Sprintf("%s에서 게이트웨이 통신을 점검합니다....\n\n%s\n\n%s",
		nodeIP, gwResult.RawOutput, gwMessage)
	emit(pub, nodeIP, "gateway_result", gwStatus, gwFullMsg, gwResult, 68)

	// Stage 4: External ICMP
	extResult, err := CheckPing(externalIP)
	extStatus, extMessage := derivePingStatus(extResult, err)
	extFullMsg := fmt.Sprintf("%s에서 외부 인터넷 연결을 점검합니다....\n\n%s\n\n%s",
		nodeIP, extResult.RawOutput, extMessage)
	emit(pub, nodeIP, "external_result", extStatus, extFullMsg, extResult, 68)

	// Stage 5: Complete
	emit(pub, nodeIP, "complete", models.StatusInfo,
		fmt.Sprintf("%s의 네트워크 점검이 완료되었습니다.", nodeIP), nil, 32)
}

// detectNodeIP reads NODE_IP injected by the Downward API (status.hostIP).
func detectNodeIP() (string, error) {
	if ip := os.Getenv("NODE_IP"); ip != "" {
		return ip, nil
	}
	return "", fmt.Errorf("NODE_IP 환경변수가 설정되지 않았습니다. Job manifest의 Downward API 설정을 확인하세요")
}

func emit(pub mq.Publisher, nodeIP, stage, status, message string, data interface{}, bannerWidth int, extraLines ...string) {
	lines := []string{message}
	if len(extraLines) > 0 {
		lines = append(lines, extraLines...)
	}
	printBanner(lines, bannerWidth)

	msg := models.DiagMessage{
		UserID:    1,
		SSEType:   "network_test",
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

func printBanner(lines []string, width int) {
	sep := strings.Repeat("=", width)
	fmt.Println(sep)
	for _, line := range lines {
		fmt.Println(line)
	}
	fmt.Println(sep)
}

func deriveARPStatus(result models.ARPResult, err error) (string, string) {
	if err != nil {
		return models.StatusError, fmt.Sprintf("ARP 조회 실행 오류: %v", err)
	}
	switch result.State {
	case "REACHABLE":
		return models.StatusOK, fmt.Sprintf("%s ARP REACHABLE (MAC: %s)", result.TargetIP, result.MAC)
	case "STALE", "DELAY", "PROBE":
		return models.StatusWarning, fmt.Sprintf("%s ARP %s (갱신 필요, MAC: %s)", result.TargetIP, result.State, result.MAC)
	default:
		return models.StatusError, fmt.Sprintf("%s ARP 항목 없음 또는 %s", result.TargetIP, result.State)
	}
}

func derivePingStatus(result models.PingResult, err error) (string, string) {
	if err != nil {
		return models.StatusError, fmt.Sprintf("ping 실행 오류: %v", err)
	}
	switch {
	case result.PacketLoss >= 100.0:
		return models.StatusError, fmt.Sprintf("%s 전체 패킷 손실", result.TargetIP)
	case result.PacketLoss > 0.0:
		return models.StatusWarning, fmt.Sprintf("%s 부분 패킷 손실 %.2f%% (avg RTT: %.2fms)", result.TargetIP, result.PacketLoss, result.RTTAvg)
	default:
		return models.StatusOK, fmt.Sprintf("%s 응답 정상 (avg RTT: %.2fms)", result.TargetIP, result.RTTAvg)
	}
}
