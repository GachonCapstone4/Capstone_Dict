package models

import "time"

type DiagMessage struct {
	UserID    int         `json:"user_id"`
	SSEType   string      `json:"sse_type"`
	Module    string      `json:"module"`
	NodeIP    string      `json:"node_ip"`
	Stage     string      `json:"stage"`
	Status    string      `json:"status"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

type ARPResult struct {
	TargetIP string `json:"target_ip"`
	MAC      string `json:"mac"`
	State    string `json:"state"`
}

type PingResult struct {
	TargetIP    string  `json:"target_ip"`
	Transmitted int     `json:"transmitted"`
	Received    int     `json:"received"`
	PacketLoss  float64 `json:"packet_loss_pct"`
	RTTMin      float64 `json:"rtt_min_ms"`
	RTTAvg      float64 `json:"rtt_avg_ms"`
	RTTMax      float64 `json:"rtt_max_ms"`
	RawOutput   string  `json:"raw_output,omitempty"`
}

type TunnelResult struct {
	PeerIP      string `json:"peer_ip"`
	WGOutput    string `json:"wg_output"`
	PingOutput  string `json:"ping_output"`
	RawOutput   string `json:"raw_output"`
	PeerFound   bool   `json:"peer_found"`
	PingSuccess bool   `json:"ping_success"`
}

type TracerouteResult struct {
	Target    string `json:"target"`
	RawOutput string `json:"raw_output"`
}

type IperfResult struct {
	Target    string `json:"target"`
	RawOutput string `json:"raw_output"`
}

type PortCheckResult struct {
	Target    string `json:"target"`
	Port      string `json:"port"`
	Open      bool   `json:"open"`
	RawOutput string `json:"raw_output"`
}

const (
	StatusInfo    = "info"
	StatusOK      = "ok"
	StatusWarning = "warning"
	StatusError   = "error"
)

