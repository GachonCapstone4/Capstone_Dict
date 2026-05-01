package vpn

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"capstone_network_test/internal/models"
)

const udpMaxRetry = 3

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

	//result.UDPRawOutput = runIperfUDP(target)

	return result, nil
}

func runIperfUDP(target string) string {
	var lastOut []byte
	for i := 0; i < udpMaxRetry; i++ {
		// 서버가 이전 세션을 정리할 시간 확보 (첫 시도도 포함)
		time.Sleep(3 * time.Second)

		cmd := exec.Command("iperf3", "-c", target, "-u", "-t", "5")
		out, err := cmd.CombinedOutput()
		lastOut = out
		if err == nil {
			return strings.TrimRight(string(out), "\n")
		}
		if _, ok := err.(*exec.ExitError); !ok {
			return fmt.Sprintf("iperf3 UDP 실행 실패: %v", err)
		}
	}
	return fmt.Sprintf("[UDP %d회 재시도 실패]\n%s", udpMaxRetry, strings.TrimRight(string(lastOut), "\n"))
}
