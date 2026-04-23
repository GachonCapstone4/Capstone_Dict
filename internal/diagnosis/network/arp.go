package network

import (
	"os/exec"
	"strings"

	"capstone_network_test/internal/models"
)

func CheckARP(targetIP string) (models.ARPResult, error) {
	result := models.ARPResult{TargetIP: targetIP, State: "NONE"}

	cmd := exec.Command("ip", "neigh", "show", targetIP)
	out, err := cmd.Output()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return result, nil
		}
		return result, err
	}

	output := strings.TrimSpace(string(out))
	if output == "" {
		return result, nil
	}

	mac, state := parseNeighOutput(output)
	result.MAC = mac
	result.State = state
	return result, nil
}

func parseNeighOutput(output string) (mac, state string) {
	tokens := strings.Fields(output)
	if len(tokens) == 0 {
		return "", "NONE"
	}

	for i, tok := range tokens {
		if tok == "lladdr" && i+1 < len(tokens) {
			mac = tokens[i+1]
		}
	}

	state = strings.ToUpper(tokens[len(tokens)-1])
	return mac, state
}
