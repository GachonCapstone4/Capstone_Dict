package vpn

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"capstone_network_test/internal/models"
)

func CheckIperf(target string) (models.IperfResult, error) {
	result := models.IperfResult{Target: target}
	if target == "" {
		return result, fmt.Errorf("IPERF_TARGET 환경변수가 설정되지 않았습니다")
	}

	tcpCmd := exec.Command("iperf3", "-c", target, "-M", "1380")
	tcpOut, err := tcpCmd.CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return result, fmt.Errorf("iperf3 TCP 실행 실패: %w", err)
		}
	}
	result.TCPRawOutput = strings.TrimRight(string(tcpOut), "\n")

	// iperf3 서버가 이전 세션을 정리할 시간 확보
	time.Sleep(5 * time.Second)

	udpCmd := exec.Command("iperf3", "-c", target, "-u", "-t", "5")
	udpOut, err := udpCmd.CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return result, fmt.Errorf("iperf3 UDP 실행 실패: %w", err)
		}
	}
	result.UDPRawOutput = strings.TrimRight(string(udpOut), "\n")

	return result, nil
}
