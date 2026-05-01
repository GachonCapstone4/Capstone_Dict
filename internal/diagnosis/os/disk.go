package os

import (
	"fmt"
	"os/exec"
	"strings"
)

// CheckDisk runs df -h (usage) and df -ih (inodes) and returns combined raw output.
func CheckDisk() (string, error) {
	dfOut, err := exec.Command("df", "-h").Output()
	if err != nil {
		return "", fmt.Errorf("df -h 실행 실패: %w", err)
	}

	diOut, err := exec.Command("df", "-ih").Output()
	if err != nil {
		return "", fmt.Errorf("df -ih 실행 실패: %w", err)
	}

	result := fmt.Sprintf("=== Disk Usage ===\n%s\n=== Inode Usage ===\n%s",
		strings.TrimSpace(string(dfOut)),
		strings.TrimSpace(string(diOut)),
	)
	return result, nil
}
