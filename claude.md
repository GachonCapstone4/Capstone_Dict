## 본 프로젝트는 캡스톤 프로젝트에서 종합적인 서버 진단도구 자동화 Go 프로젝트이다.
본 프로젝트의 환경은 Proxmox를 통한 가상화 환경에 3개의 VM을 k8s 클러스터로 구축하였다.

### Swiss Army Knife 패턴
본 프로젝트를 하나의 컨테이너로 묶을 것이고 해당 하는 모듈별로 기능에는
-> 물리 Netowrk 점검(proxmox 레벨)
-> OS 점검(RockyLinux)
-> k8s 클러스터 점검
-> VPN 상태 점검 (하이브리드 아키텍쳐임)
-> RabbitMQ 상태점검

총 5가지의 진단 테스트 모듈을 구현한다..
각 모듈들은 하나의 컨테이너지만 manifest에서 정의한 job 의 실행시에 manifest에 정의된
해당하는 모듈의 기능만 수행할수 있도록 분리되어야한다.
ex) network-manifest에서
spec:
template:
spec:
containers:
- name: job-container
image: my-image:latest
args: ["network"]  # 여기서 실행할 모듈 지정

이런식으로 특정 모듈만 실행할 수 있어야한다.
본 스펙은 전체적인 진단도구의 형태를 설계하는 것으로 본 스펙을 구체적 구현의 기준으로 잡으면 안된다.
job manifest의 정의는 본 프로젝트의 담당이 아니다.
즉 Go를 통한 서버 진단 프로그램의 개발 및 Dockerfile 작성 및 CI 프로세스까지가 본 담당이다.

본 진단 도구 모듈의 실행 결과를 json 형태로 RabbiMQ x-sse-fanout 큐에 publish 하여야 한다.
구체적인 모듈별 기능과 publish 타이밍 및 형태에 대해서는 추가 spec에서 기술한다.
절대 본 spec 을 구체적 구현의 기준으로 삼으면 안되고 설계의 방향으로 잡아야한다.
기본적으로 network와 os 테스트는 모든 노드에서 실행하는 것을 기준으로 삼는다.

/capstone_test_tools
├── cmd/
│   └── diag-tool/             # 메인 진단 도구 진입점
│       └── main.go           # CLI 서브커맨드 라우팅 (Cobra/Flag)
├── internal/                 # 외부에서 import 불가능한 프로젝트 핵심 로직
│   ├── app/                  # CLI 명령어 정의 및 실행 흐름 제어
│   │   ├── root.go
│   │   ├── network_cmd.go    # "network" 모듈 실행부
│   │   ├── os_cmd.go         # "os" 모듈 실행부
│   │   ├── k8s_cmd.go
│   │   ├── vpn_cmd.go
│   │   └── rabbitmq_cmd.go
│   ├── diagnosis/            # 실제 진단 로직 (Core Modules)
│   │   ├── network/          # Proxmox 레벨 네트워크 점검 로직
│   │   ├── os/               # RockyLinux 점검 로직
│   │   ├── k8s/              # 클러스터 점검 (client-go 활용)
│   │   ├── vpn/              # VPN 상태 점검 로직
│   │   └── rabbitmq/         # RabbitMQ 자체 상태 점검 로직
│   ├── mq/                   # RabbitMQ Client (Publisher 로직)
│   │   └── publisher.go      # JSON 메시지 발행 및 x-sse-fanout 설정
│   └── models/               # 공통 데이터 구조 (JSON 스키마 정의)
│       └── result.go         # 진단 결과 공통 인터페이스/구조체
├── configs/                  # 설정 파일 (YAML, ENV 등)
├── scripts/                  # CI/CD 및 배포 스크립트
├── Dockerfile                # 멀티 스테이지 빌드 정의
├── go.mod
└── go.sum

구조 제안 예시이고 필요시 해당 구조에서 수정가능.
현재 까지는 점검 내용을 저장할 필요는 없다.

rabbitmq-headless.rabbitmq = Rabbitmq의 내부해석 서비스 주소

claude.md 읽은후엔 반드시 system.md를 읽을 것.