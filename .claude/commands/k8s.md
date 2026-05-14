## k8s node 상태점검

skill.md 반드시 먼저 읽을것.
internal/dignosis/k8s 디렉터리안에 메인 파일을 작성한다.
network/os와는 달리 daemonset으로 모든 노드에서 실행될 필요 없고, controlplane에서 하나의 job만 동작하면 된다.
위 점을 제외한 기본적인 점검 방법과 flow는 네트워크와 동일하다.
summary.md의 기술된 내용과 변경점이 몇몇 존재하고 최신화 되지 못했다.
network 점검도구가 실제로 사용되고 있으며 즉 markdown 보다 실제 internal 내부의 code가 신뢰성이 높다.
네트워크와 동일하게 점검후 sse pod로 user-id / sse-type / data 로 이루어진 값을 보내야한다.
아마 기존에는 network 폴더에서 훨씬 많은 key를 보내고 있을 텐데 실질적으로 sse 에 서 받아볼 필요가 있는 로그는
위 3개로 이루어지면된다.
user-id = 1 (관리자의 id)
sse-type = k8s
data = 실제 관리자 웹에서 받아볼 시스템 점검 결과, 즉 본 기능에서 확인하고자 하는것.
data에는 아래 점검 항목결과의 Linux cli 에서의 실행결과인 raw 레벨 출력이 담기면된다.

### 점검항목

node 컨디션점검
kubectl get nodes 
kubectl descirbe node (k8s-worker-1 / k8s-worker-2)

pod status (kubectl get pod -A)

### 추가사항 5/14
pod 의 상태가 if not Running pod만 골라서 출력하기
현재 실행되는 잡 자체에는 당연히 kubectl 명령어가 없을 것이기 때문에 k8s api 호출을 해야하며 호출결과 반환되는 output을
그대로 받아와서 data에 넣는 방향

위 kubectl 명령어의 결과인 raw_data(시스템 output 데이터)가 publish 할때 큐에 data로 담겨야한다.

