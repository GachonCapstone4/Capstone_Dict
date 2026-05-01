package vpn

import (
	"fmt"
	"os/exec"
	"strings"

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
