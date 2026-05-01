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

	cmd := exec.Command("iperf3", "-c", target, "-u", "-t", "5")
	out, err := cmd.Output()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return result, fmt.Errorf("iperf3 바이너리 실행 실패: %w", err)
		}
	}
	result.RawOutput = strings.TrimRight(string(out), "\n")
	return result, nil
}
