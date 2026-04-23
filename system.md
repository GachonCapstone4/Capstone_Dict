# IDC Infrastructure Overview

이 문서는 홈 IDC(Proxmox 기반 온프레미스 인프라)의 전체 구조를 설명합니다.
다른 프로젝트에서 이 인프라와 연계 작업 시 참고하세요.

---

## 하이퍼바이저

| 항목 | 값 |
|---|---|
| 플랫폼 | Proxmox VE |
| 호스트명 | `suhansrv` |
| 관리 주소 | `https://192.168.100.210:8006` |
| 인증 | API Token (`admin@pve!terraform-token`) |
| IaC 도구 | Terraform (`bpg/proxmox` provider ~0.66) |

---

## 네트워크 구조

pfSense(VM 200)가 전체 서브넷의 게이트웨이/방화벽 역할을 담당합니다.

| 브릿지 | 역할 | 서브넷 |
|---|---|---|
| `vmbr0` | WAN / 물리 네트워크 (관리망 포함) | `192.168.100.0/24` |
| `vmbr1` | DMZ | — |
| `vmbr2` | App / Kubernetes Cluster 서브넷 | `192.168.2.0/24` |
| `vmbr3` | DB 서브넷 | `192.168.3.0/24` |

---

## VM / 컨테이너 목록

### VM 200 — pfSense Gateway (방화벽/라우터)

| 항목 | 값 |
|---|---|
| VM ID | 200 |
| OS | pfSense (FreeBSD 기반) |
| BIOS | OVMF (UEFI) / q35 머신 |
| CPU | 2 vCPU (host passthrough, AES-NI) |
| Memory | 2 GB |
| Disk | 20 GB (local-lvm) |
| 네트워크 | vmbr0(WAN) / vmbr1(DMZ) / vmbr2(App) / vmbr3(DB) |
| 부팅 순서 | 1 (가장 먼저 기동) |
| 역할 | 전체 서브넷 게이트웨이 + 방화벽 + VPN |

---

### LXC 201 — Tailscale Subnet Router

| 항목 | 값 |
|---|---|
| VM ID | 201 |
| OS | Ubuntu 24.04 LXC |
| IP | `192.168.100.211/24` (vmbr0) |
| Gateway | `192.168.100.254` |
| Disk | 8 GB (local-lvm) |
| 역할 | Tailscale 서브넷 라우터 (원격 접근용) |
| 비고 | `/dev/net/tun` 마운트 및 Tailscale 설치는 `script/init.sh`로 수동 설정 |

---

### VM 300 — Kubernetes Control Plane

| 항목 | 값 |
|---|---|
| VM ID | 300 |
| OS | Rocky Linux 10 (템플릿 VM 9001 클론) |
| IP | `192.168.2.10/24` |
| Gateway | `192.168.2.1` (pfSense vmbr2) |
| CPU | 2 vCPU |
| Memory | 6 GB |
| Disk | 20 GB (local-lvm) |
| 네트워크 | vmbr2 (App/Cluster 서브넷) |
| 부팅 순서 | 2 (pfSense 기동 후 30초 대기) |
| 역할 | Kubernetes Control Plane 노드 |

---

### VM 301, 302 — Kubernetes Worker Nodes

| 항목 | 값 |
|---|---|
| VM ID | 301 (worker-1) / 302 (worker-2) |
| OS | Rocky Linux 10 (템플릿 VM 9001 클론) |
| IP | `192.168.2.20/24` (worker-1) / `192.168.2.30/24` (worker-2) |
| Gateway | `192.168.2.1` (pfSense vmbr2) |
| CPU | 4 vCPU |
| Memory | 8 GB |
| Disk | 40 GB (local-lvm) |
| 네트워크 | vmbr2 (App/Cluster 서브넷) |
| 부팅 순서 | 3 (control-plane 기동 후 60초 대기) |
| 역할 | Kubernetes Worker 노드 |

---

### VM 400 — MariaDB Master

| 항목 | 값 |
|---|---|
| VM ID | 400 |
| OS | Rocky Linux 10 (템플릿 VM 9001 클론) |
| IP | `192.168.3.10/24` |
| Gateway | `192.168.3.1` (pfSense vmbr3) |
| CPU | 2 vCPU |
| Memory | 2 GB |
| Disk | 20 GB (local-lvm) |
| 네트워크 | vmbr3 (DB 서브넷) |
| 역할 | MariaDB Master 노드 |

---

### VM 401 — MariaDB Slave

| 항목 | 값 |
|---|---|
| VM ID | 401 |
| OS | Rocky Linux 10 (템플릿 VM 9001 클론) |
| IP | `192.168.3.20/24` |
| Gateway | `192.168.3.1` (pfSense vmbr3) |
| CPU | 2 vCPU |
| Memory | 2 GB |
| Disk | 20 GB (local-lvm) |
| 네트워크 | vmbr3 (DB 서브넷) |
| 역할 | MariaDB Slave (Replica) 노드 |

---

## 전체 토폴로지

```
[외부 인터넷 / 물리망 192.168.100.0/24]
         |
    [vmbr0: WAN]
         |
  ┌──────────────┐
  │  VM 200      │  ← 부팅 순서 1
  │  pfSense     │  (게이트웨이 / 방화벽)
  └──┬───┬───┬───┘
     │   │   │
  vmbr2 vmbr3 vmbr1(DMZ)
     │   │
     │   └─────────────────────────────┐
     │                                 │
  [App/Cluster 192.168.2.0/24]    [DB 192.168.3.0/24]
     │                                 │
  VM 300 Control Plane (.10)      VM 400 MariaDB Master (.10)
  VM 301 Worker-1 (.20)           VM 401 MariaDB Slave  (.20)
  VM 302 Worker-2 (.30)

[vmbr0]
  └── LXC 201 Tailscale Router (192.168.100.211)
       └── Tailscale 메시 VPN으로 원격 접근 제공
```

---

## 템플릿

| VM ID | 이름 | 설명 |
|---|---|---|
| 9001 | Rocky10-Template | Rocky Linux 10 cloud-init 템플릿. 모든 VM 클론의 베이스. qemu-guest-agent 필수 설치 상태여야 함 |

---

## 접근 정보 요약

- **Proxmox Web UI**: `https://192.168.100.210:8006` (자체 서명 인증서)
- **원격 접근**: Tailscale VPN (LXC 201 서브넷 라우터 경유)
- **VM 기본 계정**: `suhan` (패스워드는 `terraform.tfvars`의 `lxcpw` / `vmpassword`)
- **Terraform 루트 모듈**: `stage/` 디렉토리에서 실행
