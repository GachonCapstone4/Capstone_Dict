package os

import (
	"fmt"
	"os/exec"
	"strings"
)

// CheckCPU runs top -bn1 and returns the summary header lines
// (uptime/load average + task counts + CPU% + memory lines).
func CheckCPU() (string, error) {
	out, err := exec.Command("top", "-bn1").Output()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return "", fmt.Errorf("top 실행 실패: %w", err)
		}
	}

	lines := strings.Split(string(out), "\n")
	n := 6
	if len(lines) < n {
		n = len(lines)
	}
	return strings.Join(lines[:n], "\n"), nil
}
