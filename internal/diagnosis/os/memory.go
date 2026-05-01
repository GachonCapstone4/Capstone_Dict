package os

import (
	"fmt"
	"os/exec"
	"strings"
)

// CheckMemory runs free -h and returns its raw output.
func CheckMemory() (string, error) {
	out, err := exec.Command("free", "-h").Output()
	if err != nil {
		return "", fmt.Errorf("free 실행 실패: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
