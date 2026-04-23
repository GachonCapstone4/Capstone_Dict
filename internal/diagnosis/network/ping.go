package network

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"

	"capstone_network_test/internal/models"
)

var (
	rePacketStats = regexp.MustCompile(`(\d+) packets transmitted, (\d+) received, ([\d.]+)% packet loss`)
	reRTT         = regexp.MustCompile(`rtt min/avg/max/mdev = ([\d.]+)/([\d.]+)/([\d.]+)/([\d.]+) ms`)
)

func CheckPing(targetIP string) (models.PingResult, error) {
	result := models.PingResult{TargetIP: targetIP}

	cmd := exec.Command("ping", "-c", "3", "-W", "2", targetIP)
	out, err := cmd.Output()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return result, fmt.Errorf("ping 바이너리 실행 실패: %w", err)
		}
	}

	output := string(out)

	transmitted, received, loss, err := parsePacketStats(output)
	if err == nil {
		result.Transmitted = transmitted
		result.Received = received
		result.PacketLoss = loss
	}

	min, avg, max, err := parseRTTStats(output)
	if err == nil {
		result.RTTMin = min
		result.RTTAvg = avg
		result.RTTMax = max
	}

	return result, nil
}

func parsePacketStats(output string) (transmitted, received int, packetLoss float64, err error) {
	m := rePacketStats.FindStringSubmatch(output)
	if m == nil {
		return 0, 0, 0, fmt.Errorf("패킷 통계 파싱 실패")
	}
	transmitted, _ = strconv.Atoi(m[1])
	received, _ = strconv.Atoi(m[2])
	packetLoss, _ = strconv.ParseFloat(m[3], 64)
	return transmitted, received, packetLoss, nil
}

func parseRTTStats(output string) (min, avg, max float64, err error) {
	m := reRTT.FindStringSubmatch(output)
	if m == nil {
		return 0, 0, 0, fmt.Errorf("RTT 통계 파싱 실패")
	}
	min, _ = strconv.ParseFloat(m[1], 64)
	avg, _ = strconv.ParseFloat(m[2], 64)
	max, _ = strconv.ParseFloat(m[3], 64)
	return min, avg, max, nil
}
