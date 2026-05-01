package vpn

import (
	"fmt"
	"os/exec"
	"strings"

	"capstone_network_test/internal/models"
)

func CheckPortNC(target, port string) (models.PortCheckResult, error) {
	result := models.PortCheckResult{Target: target, Port: port}
	if target == "" {
		return result, fmt.Errorf("NC_TARGET 환경변수가 설정되지 않았습니다")
	}

	cmd := exec.Command("nc", "-zv", "-w", "5", target, port)
	out, err := cmd.CombinedOutput()
	result.RawOutput = strings.TrimRight(string(out), "\n")

	if err == nil {
		result.Open = true
		return result, nil
	}
	if _, ok := err.(*exec.ExitError); ok {
		// exit code 1 = 포트 닫힘/거부, 바이너리 자체 오류 아님
		result.Open = false
		return result, nil
	}
	return result, fmt.Errorf("nc 바이너리 실행 실패: %w", err)
}
