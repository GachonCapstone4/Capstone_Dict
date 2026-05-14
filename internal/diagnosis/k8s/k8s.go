package k8s

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

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

	clientset, err := newClient()
	if err != nil {
		emit(pub, "init_error", models.StatusError, fmt.Sprintf("k8s 클라이언트 초기화 실패: %v", err), nil, 68)
		return
	}

	ctx := context.Background()

	// kubectl get nodes
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		emit(pub, "nodes_result", models.StatusError, fmt.Sprintf("노드 목록 조회 실패: %v", err), nil, 68)
	} else {
		emit(pub, "nodes_result", models.StatusOK, "노드 컨디션 점검", rawData(formatNodeList(nodes)), 68)
	}

	// kubectl describe node worker1 / worker2
	for _, workerName := range []string{worker1, worker2} {
		stage := strings.ReplaceAll(workerName, "-", "_") + "_describe_result"
		node, err := clientset.CoreV1().Nodes().Get(ctx, workerName, metav1.GetOptions{})
		if err != nil {
			emit(pub, stage, models.StatusError, fmt.Sprintf("%s 노드 상세 조회 실패: %v", workerName, err), nil, 68)
		} else {
			emit(pub, stage, models.StatusOK, fmt.Sprintf("%s 노드 상세 점검", workerName), rawData(formatNodeDescribe(node)), 68)
		}
	}

	// kubectl get pod -A
	pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		emit(pub, "pods_result", models.StatusError, fmt.Sprintf("파드 목록 조회 실패: %v", err), nil, 68)
		emit(pub, "complete", models.StatusInfo, "k8s 클러스터 점검이 완료되었습니다.", nil, 32)
		return
	}

	emit(pub, "pods_result", models.StatusOK, "전체 네임스페이스 파드 상태 점검", rawData(formatPodList(pods.Items)), 68)

	// not Running 파드 필터
	notRunning := filterNotRunningPods(pods.Items)
	if len(notRunning) == 0 {
		emit(pub, "not_running_pods_result", models.StatusOK, "비정상 파드 없음 (모든 파드 Running)", nil, 68)
	} else {
		emit(pub, "not_running_pods_result", models.StatusError, "비정상 파드 감지됨", rawData(formatPodList(notRunning)), 68)
	}

	emit(pub, "complete", models.StatusInfo, "k8s 클러스터 점검이 완료되었습니다.", nil, 32)
}

func newClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("인클러스터 설정 로드 실패: %w", err)
	}
	return kubernetes.NewForConfig(config)
}

func formatNodeList(nodes *corev1.NodeList) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%-30s %-10s %-15s %-10s\n", "NAME", "STATUS", "ROLES", "VERSION"))
	for _, node := range nodes.Items {
		status := "NotReady"
		for _, cond := range node.Status.Conditions {
			if cond.Type == corev1.NodeReady && cond.Status == corev1.ConditionTrue {
				status = "Ready"
				break
			}
		}
		var roles []string
		for label := range node.Labels {
			if strings.HasPrefix(label, "node-role.kubernetes.io/") {
				roles = append(roles, strings.TrimPrefix(label, "node-role.kubernetes.io/"))
			}
		}
		sort.Strings(roles)
		roleStr := "<none>"
		if len(roles) > 0 {
			roleStr = strings.Join(roles, ",")
		}
		sb.WriteString(fmt.Sprintf("%-30s %-10s %-15s %-10s\n",
			node.Name, status, roleStr, node.Status.NodeInfo.KubeletVersion))
	}
	return sb.String()
}

func formatNodeDescribe(node *corev1.Node) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Name: %s\n", node.Name))

	var roles []string
	for label := range node.Labels {
		if strings.HasPrefix(label, "node-role.kubernetes.io/") {
			roles = append(roles, strings.TrimPrefix(label, "node-role.kubernetes.io/"))
		}
	}
	sort.Strings(roles)
	sb.WriteString(fmt.Sprintf("Roles: %s\n", strings.Join(roles, ",")))

	sb.WriteString("Addresses:\n")
	for _, addr := range node.Status.Addresses {
		sb.WriteString(fmt.Sprintf("  %s: %s\n", addr.Type, addr.Address))
	}

	sb.WriteString("Conditions:\n")
	sb.WriteString(fmt.Sprintf("  %-20s %-8s %s\n", "Type", "Status", "Message"))
	for _, cond := range node.Status.Conditions {
		msg := cond.Message
		if len(msg) > 60 {
			msg = msg[:57] + "..."
		}
		sb.WriteString(fmt.Sprintf("  %-20s %-8s %s\n", cond.Type, cond.Status, msg))
	}

	sb.WriteString(fmt.Sprintf("Capacity:\n  cpu: %s\n  memory: %s\n",
		node.Status.Capacity.Cpu(), node.Status.Capacity.Memory()))
	sb.WriteString(fmt.Sprintf("Allocatable:\n  cpu: %s\n  memory: %s\n",
		node.Status.Allocatable.Cpu(), node.Status.Allocatable.Memory()))

	sb.WriteString(fmt.Sprintf("OS: %s\n", node.Status.NodeInfo.OSImage))
	sb.WriteString(fmt.Sprintf("KubeletVersion: %s\n", node.Status.NodeInfo.KubeletVersion))
	sb.WriteString(fmt.Sprintf("ContainerRuntime: %s\n", node.Status.NodeInfo.ContainerRuntimeVersion))

	return sb.String()
}

func formatPodList(pods []corev1.Pod) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%-20s %-45s %-8s %-12s %-10s\n", "NAMESPACE", "NAME", "READY", "STATUS", "RESTARTS"))
	for _, pod := range pods {
		ready, total, restarts := 0, len(pod.Spec.Containers), 0
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.Ready {
				ready++
			}
			restarts += int(cs.RestartCount)
		}
		sb.WriteString(fmt.Sprintf("%-20s %-45s %-8s %-12s %-10d\n",
			pod.Namespace, pod.Name,
			fmt.Sprintf("%d/%d", ready, total),
			string(pod.Status.Phase),
			restarts))
	}
	return sb.String()
}

func filterNotRunningPods(pods []corev1.Pod) []corev1.Pod {
	var result []corev1.Pod
	for _, pod := range pods {
		if pod.Status.Phase != corev1.PodRunning {
			result = append(result, pod)
		}
	}
	return result
}

func rawData(raw string) interface{} {
	return map[string]string{"raw_output": raw}
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
