package os

import (
	"fmt"
	"os/exec"
	"strings"
)

// CheckProcess runs ps aux and filters for zombie (Z) and uninterruptible (D) state processes.
// Returns the header + matching lines, or a clean message if none found.
func CheckProcess() (string, error) {
	out, err := exec.Command("ps", "aux").Output()
	if err != nil {
		return "", fmt.Errorf("ps aux 실행 실패: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 {
		return "ps aux 출력 없음", nil
	}

	header := lines[0]
	var abnormal []string
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		// ps aux STAT column is index 7
		if len(fields) < 8 {
			continue
		}
		stat := fields[7]
		if strings.HasPrefix(stat, "Z") || strings.HasPrefix(stat, "D") {
			abnormal = append(abnormal, line)
		}
	}

	if len(abnormal) == 0 {
		return "좀비 및 비정상 프로세스 없음 (Z/D 상태)", nil
	}

	return fmt.Sprintf("비정상 프로세스 %d개 발견\n%s\n%s",
		len(abnormal), header, strings.Join(abnormal, "\n")), nil
}
