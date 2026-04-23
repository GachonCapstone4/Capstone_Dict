# capstone_test_tools — Network 모듈 코드 요약

## 프로젝트 개요

Proxmox 기반 홈 IDC(k8s 3-node 클러스터)의 서버 진단 자동화 도구.
k8s Job 형태로 실행되며, 진단 결과를 JSON으로 RabbitMQ에 publish한다.

현재 구현: **Network 모듈** (L2 ARP → L3 게이트웨이 ICMP → 외부 ICMP)

---

## 디렉터리 구조

```
capstone_test_tools/
├── cmd/diag-tool/main.go                   # 바이너리 진입점
├── internal/
│   ├── app/
│   │   ├── root.go                         # Cobra 루트 커맨드
│   │   └── network_cmd.go                  # "network" 서브커맨드 등록
│   ├── diagnosis/
│   │   └── network/
│   │       ├── network.go                  # 5단계 오케스트레이터
│   │       ├── arp.go                      # ARP 테이블 점검 로직
│   │       └── ping.go                     # ICMP ping 점검 로직
│   ├── mq/
│   │   └── publisher.go                    # RabbitMQ AMQP 퍼블리셔
│   └── models/
│       └── result.go                       # 공통 JSON 구조체
├── go.mod
└── go.sum
```

---

## 실행 방법

```bash
# 빌드
go build ./cmd/diag-tool/

# 실행 (RabbitMQ 연결 없이도 stdout 출력은 정상 동작)
./diag-tool network

# 환경변수 지정
RABBITMQ_URL=amqp://user:pass@rabbitmq-headless.rabbitmq:5672/ ./diag-tool network
NODE_IP=192.168.2.10 ./diag-tool network   # 인터페이스 자동감지 실패 시 fallback
```

---

## 모듈별 역할

### cmd/diag-tool/main.go
`app.Execute()` 호출 후 에러 시 `os.Exit(1)`. 비즈니스 로직 없음.

### internal/app/root.go
Cobra 루트 커맨드 정의. `Execute()` 함수 export.
서브커맨드는 각자 `init()`에서 `rootCmd.AddCommand()`로 등록 → root.go 수정 없이 모듈 추가 가능.

### internal/app/network_cmd.go
`network` 서브커맨드 등록.
`mq.NewPublisher()` 연결 실패 시 경고만 출력하고 `noopPublisher`로 계속 진행.
`defer pub.Close()`로 Job 종료 시 연결 정리.

### internal/models/result.go
전 모듈 공통으로 사용하는 JSON 구조체.

| 타입 | 용도 |
|------|------|
| `DiagMessage` | RabbitMQ 전송 단위. module/node_ip/stage/status/message/data/timestamp 포함 |
| `ARPResult` | ARP 점검 결과. target_ip, MAC, State(REACHABLE/STALE/FAILED/NONE) |
| `PingResult` | Ping 점검 결과. transmitted/received/packet_loss_pct/rtt_min/avg/max |

Status 상수: `info` / `ok` / `warning` / `error`

### internal/mq/publisher.go
`Publisher` 인터페이스(`Publish`, `Close`)를 통해 live/noop 구현체를 추상화.

| 구현체 | 조건 |
|--------|------|
| `amqpPublisher` | RabbitMQ 연결 성공 시 |
| `noopPublisher` | 연결 실패 시 fallback. Publish()는 no-op |

- Exchange `x-sse-fanout`은 이미 생성되어 있으므로 **선언하지 않고 publish만 수행**
- routing key = `""` (fanout은 routing key 무시)
- k8s Job 특성상 재연결 로직 없음 — connect → publish N회 → close

### internal/diagnosis/network/network.go
5단계 실행 오케스트레이터.

**노드 IP 감지 (`detectNodeIP`)**
1. 활성화된 non-loopback 인터페이스에서 `192.168.2.x` IP 탐색
2. 없으면 `NODE_IP` 환경변수 사용
3. 둘 다 없으면 error emit 후 종료

**ARP Ring 매핑**
```
192.168.2.10 → 192.168.2.20 → 192.168.2.30 → 192.168.2.10
```
자신의 IP로 다음 노드를 결정해 해당 노드의 ARP 항목 점검.

**`emit()` 패턴**
모든 이벤트를 동일한 함수로 처리: stdout 배너 출력 → `DiagMessage` 생성 → RabbitMQ publish.
publish 실패는 로그만 출력하고 진단 계속 진행.

**5단계 흐름**

| 단계 | stage | 설명 |
|------|-------|------|
| 1 | `start` | 점검 시작 알림 (배너 폭 32) |
| 2a | `arp_start` | ARP 점검 시작 알림 (배너 폭 68) |
| 2b | `arp_result` | ARP 점검 결과 |
| 3a | `gateway_start` | 게이트웨이(192.168.2.1) ping 시작 |
| 3b | `gateway_result` | 게이트웨이 ping 결과 |
| 4a | `external_start` | 외부(8.8.8.8) ping 시작 |
| 4b | `external_result` | 외부 ping 결과 |
| 5 | `complete` | 점검 완료 (배너 폭 32) |

**Status 판정 기준**

| 조건 | status |
|------|--------|
| ARP State == REACHABLE | `ok` |
| ARP State == STALE / DELAY / PROBE | `warning` |
| ARP State == FAILED / NONE / 기타 | `error` |
| Ping 0% 손실 | `ok` |
| Ping 0% < 손실 < 100% | `warning` |
| Ping 100% 손실 | `error` |
| exec 오류 (바이너리 없음) | `error` |

### internal/diagnosis/network/arp.go
`CheckARP(targetIP)` — `ip neigh show <targetIP>` 실행 후 파싱.
- 출력 없음 → `State: "NONE"`, 오류 없음
- `ExitError` (명령은 실행됐지만 항목 없음) → 오류 아님
- `lladdr` 토큰 다음이 MAC, 마지막 토큰이 State

### internal/diagnosis/network/ping.go
`CheckPing(targetIP)` — `ping -c 3 -W 2 <targetIP>` 실행 후 파싱.
- `ExitError` (패킷 손실로 exit 1) → 오류 아님, 출력 파싱 계속
- 그 외 오류 → 바이너리 없음, error 반환
- 정규식으로 패킷 통계(전송/수신/손실%) 및 RTT(min/avg/max) 추출

---

## RabbitMQ 메시지 예시

```json
{
  "module": "network",
  "node_ip": "192.168.2.10",
  "stage": "arp_result",
  "status": "ok",
  "message": "192.168.2.20 ARP REACHABLE (MAC: aa:bb:cc:dd:ee:ff)",
  "data": {
    "target_ip": "192.168.2.20",
    "mac": "aa:bb:cc:dd:ee:ff",
    "state": "REACHABLE"
  },
  "timestamp": "2026-04-23T10:00:02Z"
}
```

---

## 환경변수

| 변수 | 기본값 | 설명 |
|------|--------|------|
| `RABBITMQ_URL` | `amqp://guest:guest@rabbitmq-headless.rabbitmq:5672/` | AMQP 연결 URL |
| `NODE_IP` | (자동 감지) | 192.168.2.x 인터페이스 없을 때 fallback |

---

## 컨테이너 요구사항

| 패키지 | 필요 이유 |
|--------|-----------|
| `iproute2` | `ip neigh show` 명령 제공 |
| `iputils` | `ping` (setuid 또는 `NET_RAW` capability 필요) |

Go 바이너리는 `CGO_ENABLED=0 GOOS=linux`로 정적 빌드 권장 (멀티 스테이지 Dockerfile).
