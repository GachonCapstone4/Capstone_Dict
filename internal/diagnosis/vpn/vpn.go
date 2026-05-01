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

const (
	moduleName  = "vpn"
	defaultPort = "443"
)

func Run(pub mq.Publisher) {
	nodeIP := os.Getenv("NODE_IP")
	if nodeIP == "" {
		nodeIP = "unknown"
	}

	vpnPeerIP := os.Getenv("VPN_PEER_IP")
	tracerouteTarget := os.Getenv("TRACEROUTE_TARGET")
	iperfTarget := os.Getenv("IPERF_TARGET")
	ncTarget := os.Getenv("NC_TARGET")
	ncPort := os.Getenv("NC_PORT")
	if ncPort == "" {
		ncPort = defaultPort
	}

	// Stage 1: Start
	emit(pub, nodeIP, "start", models.StatusInfo, "VPN 상태 점검을 시작합니다...", nil)

	// Stage 2: L3 Tunnel (wg show + ping -c 5)
	emit(pub, nodeIP, "tunnel_start", models.StatusInfo,
		fmt.Sprintf("L3 터널 상태를 점검합니다 (peer: %s)...", vpnPeerIP), nil)
	tunnelResult, err := CheckTunnel(vpnPeerIP)
	tStatus, tMessage := deriveTunnelStatus(tunnelResult, err)
	emit(pub, nodeIP, "tunnel_result", tStatus, tMessage, tunnelResult)

	// Stage 3: Path Validation (traceroute)
	emit(pub, nodeIP, "traceroute_start", models.StatusInfo,
		fmt.Sprintf("경로 검증을 시작합니다 (target: %s)...", tracerouteTarget), nil)
	traceResult, err := CheckTraceroute(tracerouteTarget)
	trStatus, trMessage := deriveTracerouteStatus(traceResult, err)
	emit(pub, nodeIP, "traceroute_result", trStatus, trMessage, traceResult)

	// Stage 4: Performance Quality (iperf3 -c [IP] -u)
	emit(pub, nodeIP, "iperf_start", models.StatusInfo,
		fmt.Sprintf("성능 품질을 측정합니다 (target: %s)...", iperfTarget), nil)
	iperfResult, err := CheckIperf(iperfTarget)
	iStatus, iMessage := deriveIperfStatus(iperfResult, err)
	emit(pub, nodeIP, "iperf_result", iStatus, iMessage, iperfResult)

	// Stage 5: Security Policy (nc -zv [IP] [Port])
	emit(pub, nodeIP, "portcheck_start", models.StatusInfo,
		fmt.Sprintf("보안 정책을 점검합니다 (%s:%s)...", ncTarget, ncPort), nil)
	ncResult, err := CheckPortNC(ncTarget, ncPort)
	nStatus, nMessage := deriveNCStatus(ncResult, err)
	emit(pub, nodeIP, "portcheck_result", nStatus, nMessage, ncResult)

	// Stage 6: Complete
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

func deriveTunnelStatus(r models.TunnelResult, err error) (string, string) {
	if err != nil {
		return models.StatusError, fmt.Sprintf("L3 터널 점검 실행 오류: %v", err)
	}
	if !r.PeerFound {
		return models.StatusError, "WireGuard peer 정보를 찾을 수 없습니다"
	}
	if !r.PingSuccess {
		return models.StatusWarning, fmt.Sprintf("WireGuard peer 발견됨, 그러나 ping 응답 없음 (peer: %s)", r.PeerIP)
	}
	return models.StatusOK, fmt.Sprintf("L3 터널 정상 (peer: %s)", r.PeerIP)
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
	if r.RawOutput == "" {
		return models.StatusError, "iperf3 출력 없음"
	}
	return models.StatusOK, fmt.Sprintf("대역폭 측정 완료 (target: %s)", r.Target)
}

func deriveNCStatus(r models.PortCheckResult, err error) (string, string) {
	if err != nil {
		return models.StatusError, fmt.Sprintf("nc 실행 오류: %v", err)
	}
	if r.Open {
		return models.StatusOK, fmt.Sprintf("%s:%s 포트 열림", r.Target, r.Port)
	}
	return models.StatusError, fmt.Sprintf("%s:%s 포트 닫힘 또는 접근 불가", r.Target, r.Port)
}
