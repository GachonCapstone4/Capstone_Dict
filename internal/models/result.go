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

const (
	StatusInfo    = "info"
	StatusOK      = "ok"
	StatusWarning = "warning"
	StatusError   = "error"
)

