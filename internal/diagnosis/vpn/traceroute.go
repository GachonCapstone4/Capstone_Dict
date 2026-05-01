package vpn

import (
	"fmt"
	"os/exec"
	"strings"

	"capstone_network_test/internal/models"
)

func CheckTraceroute(target string) (models.TracerouteResult, error) {
	result := models.TracerouteResult{Target: target}
	if target == "" {
		return result, fmt.Errorf("TRACEROUTE_TARGET 환경변수가 설정되지 않았습니다")
	}

	cmd := exec.Command("traceroute", "-n", "-m", "20", target)
	out, err := cmd.Output()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return result, fmt.Errorf("traceroute 바이너리 실행 실패: %w", err)
		}
	}
	result.RawOutput = strings.TrimRight(string(out), "\n")
	return result, nil
}
