package os

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"capstone_network_test/internal/models"
	"capstone_network_test/internal/mq"
)

func Run(pub mq.Publisher) {
	nodeIP, err := detectNodeIP()
	if err != nil {
		log.Printf("[os] 노드 IP 감지 실패: %v", err)
		emit(pub, "unknown", "start", models.StatusError,
			fmt.Sprintf("노드 IP 감지 실패: %v", err), "", 32)
		return
	}

	// Stage 1: Start
	emit(pub, nodeIP, "start", models.StatusInfo,
		fmt.Sprintf("%s의 OS 점검을 시작합니다....", nodeIP), "", 32)

	// Stage 2: CPU / Load Average
	cpuOut, err := CheckCPU()
	cpuStatus, cpuMsg := statusFromErr(err, "CPU/Load Average 점검 완료", "CPU/Load Average 점검 오류")
	emit(pub, nodeIP, "cpu_result", cpuStatus,
		fmt.Sprintf("%s CPU/Load Average 점검 결과\n\n%s", nodeIP, cpuMsg), cpuOut, 68)

	// Stage 3: Memory
	memOut, err := CheckMemory()
	memStatus, memMsg := statusFromErr(err, "메모리 점검 완료", "메모리 점검 오류")
	emit(pub, nodeIP, "memory_result", memStatus,
		fmt.Sprintf("%s 메모리 점검 결과\n\n%s", nodeIP, memMsg), memOut, 68)

	// Stage 4: Disk
	diskOut, err := CheckDisk()
	diskStatus, diskMsg := statusFromErr(err, "디스크 점검 완료", "디스크 점검 오류")
	emit(pub, nodeIP, "disk_result", diskStatus,
		fmt.Sprintf("%s 디스크 파티션/inode 점검 결과\n\n%s", nodeIP, diskMsg), diskOut, 68)

	// Stage 5: Process
	procOut, err := CheckProcess()
	procStatus, procMsg := statusFromErr(err, "프로세스 점검 완료", "프로세스 점검 오류")
	emit(pub, nodeIP, "process_result", procStatus,
		fmt.Sprintf("%s 프로세스 점검 결과\n\n%s", nodeIP, procMsg), procOut, 68)

	// Stage 6: Complete
	emit(pub, nodeIP, "complete", models.StatusInfo,
		fmt.Sprintf("%s의 OS 점검이 완료되었습니다.", nodeIP), "", 32)
}

func detectNodeIP() (string, error) {
	if ip := os.Getenv("NODE_IP"); ip != "" {
		return ip, nil
	}
	return "", fmt.Errorf("NODE_IP 환경변수가 설정되지 않았습니다. Job manifest의 Downward API 설정을 확인하세요")
}

func statusFromErr(err error, okMsg, errPrefix string) (string, string) {
	if err != nil {
		return models.StatusError, fmt.Sprintf("%s: %v", errPrefix, err)
	}
	return models.StatusOK, okMsg
}

// emit publishes a DiagMessage whose Data field carries {"raw_output": rawOutput}
// so that the SSE hub's extractText can pick it up via data["raw_output"].
func emit(pub mq.Publisher, nodeIP, stage, status, message, rawOutput string, bannerWidth int) {
	printBanner([]string{message}, bannerWidth)

	var data interface{}
	if rawOutput != "" {
		data = map[string]string{"raw_output": rawOutput}
	}

	msg := models.DiagMessage{
		UserID:    1,
		SSEType:   "os",
		Module:    "os",
		NodeIP:    nodeIP,
		Stage:     stage,
		Status:    status,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().UTC(),
	}

	if err := pub.Publish(msg); err != nil {
		log.Printf("[MQ] publish 실패 (stage=%s, node=%s): %v", stage, nodeIP, err)
	}
}

func printBanner(lines []string, width int) {
	sep := strings.Repeat("=", width)
	fmt.Println(sep)
	for _, line := range lines {
		fmt.Println(line)
	}
	fmt.Println(sep)
}
