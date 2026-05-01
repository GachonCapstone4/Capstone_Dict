package vpn

import (
	"fmt"
	"os/exec"
	"strings"

	"capstone_network_test/internal/models"
)

func CheckTunnel(peerIP string) (models.TunnelResult, error) {
	result := models.TunnelResult{PeerIP: peerIP}

	wgCmd := exec.Command("wg", "show")
	wgOut, err := wgCmd.Output()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return result, fmt.Errorf("wg 바이너리 실행 실패: %w", err)
		}
	}
	result.WGOutput = strings.TrimRight(string(wgOut), "\n")
	result.PeerFound = strings.Contains(result.WGOutput, "peer:")

	if peerIP != "" {
		pingCmd := exec.Command("ping", "-c", "5", "-W", "2", peerIP)
		pingOut, _ := pingCmd.Output()
		result.PingOutput = strings.TrimRight(string(pingOut), "\n")
		result.PingSuccess = strings.Contains(result.PingOutput, " 0% packet loss") ||
			(!strings.Contains(result.PingOutput, "100% packet loss") &&
				strings.Contains(result.PingOutput, "bytes from"))
	}

	result.RawOutput = fmt.Sprintf("=== wg show ===\n%s\n\n=== ping -c 5 %s ===\n%s",
		result.WGOutput, peerIP, result.PingOutput)
	return result, nil
}
