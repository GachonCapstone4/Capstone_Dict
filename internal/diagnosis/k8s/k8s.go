package k8s

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"capstone_network_test/internal/models"
	"capstone_network_test/internal/mq"
)

const moduleName = "k8s"

func Run(pub mq.Publisher) {
	worker1 := os.Getenv("K8S_WORKER1")
	if worker1 == "" {
		worker1 = "k8s-worker-1"
	}
	worker2 := os.Getenv("K8S_WORKER2")
	if worker2 == "" {
		worker2 = "k8s-worker-2"
	}

	emit(pub, "start", models.StatusInfo, "k8s 클러스터 점검을 시작합니다...", nil, 32)

	raw, err := runKubectl("get", "nodes")
	msg := "노드 컨디션 점검"
	if err != nil {
		msg = fmt.Sprintf("kubectl get nodes 실패: %v", err)
	}
	emit(pub, "nodes_result", deriveStatus(err), msg, rawData(raw), 68)

	raw, err = runKubectl("describe", "node", worker1)
	msg = fmt.Sprintf("%s 노드 상세 점검", worker1)
	if err != nil {
		msg = fmt.Sprintf("kubectl describe node %s 실패: %v", worker1, err)
	}
	emit(pub, "worker1_describe_result", deriveStatus(err), msg, rawData(raw), 68)

	raw, err = runKubectl("describe", "node", worker2)
	msg = fmt.Sprintf("%s 노드 상세 점검", worker2)
	if err != nil {
		msg = fmt.Sprintf("kubectl describe node %s 실패: %v", worker2, err)
	}
	emit(pub, "worker2_describe_result", deriveStatus(err), msg, rawData(raw), 68)

	raw, err = runKubectl("get", "pod", "-A")
	msg = "전체 네임스페이스 파드 상태 점검"
	if err != nil {
		msg = fmt.Sprintf("kubectl get pod -A 실패: %v", err)
	}
	emit(pub, "pods_result", deriveStatus(err), msg, rawData(raw), 68)

	notRunning := filterNotRunningPods(raw)
	if notRunning == "" {
		emit(pub, "not_running_pods_result", models.StatusOK, "비정상 파드 없음 (모든 파드 Running)", nil, 68)
	} else {
		emit(pub, "not_running_pods_result", models.StatusError, "비정상 파드 감지됨", rawData(notRunning), 68)
	}

	emit(pub, "complete", models.StatusInfo, "k8s 클러스터 점검이 완료되었습니다.", nil, 32)
}

// filterNotRunningPods parses `kubectl get pod -A` output and returns only non-Running pod lines.
// Returns empty string if all pods are Running.
func filterNotRunningPods(raw string) string {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	if len(lines) == 0 {
		return ""
	}

	var result []string
	result = append(result, lines[0]) // header

	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		// kubectl get pod -A: NAMESPACE NAME READY STATUS RESTARTS AGE
		if len(fields) < 4 {
			continue
		}
		status := fields[3]
		if status != "Running" {
			result = append(result, line)
		}
	}

	if len(result) <= 1 {
		return ""
	}
	return strings.Join(result, "\n")
}

func runKubectl(args ...string) (string, error) {
	out, err := exec.Command("kubectl", args...).CombinedOutput()
	return string(out), err
}

func rawData(raw string) interface{} {
	if raw == "" {
		return nil
	}
	return map[string]string{"raw_output": raw}
}

func deriveStatus(err error) string {
	if err != nil {
		return models.StatusError
	}
	return models.StatusOK
}

func emit(pub mq.Publisher, stage, status, message string, data interface{}, bannerWidth int) {
	printBanner(message, bannerWidth)

	msg := models.DiagMessage{
		UserID:    1,
		SSEType:   "k8s",
		Module:    moduleName,
		Stage:     stage,
		Status:    status,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().UTC(),
	}

	if err := pub.Publish(msg); err != nil {
		log.Printf("[MQ] publish 실패 (stage=%s): %v", stage, err)
	}
}

func printBanner(message string, width int) {
	sep := strings.Repeat("=", width)
	fmt.Println(sep)
	fmt.Println(message)
	fmt.Println(sep)
}
