package main

import "time"

// Standards for the result by launching a Missile
type Harm struct {
    startTime time.Time `json:"start_time"`
    endTime time.Time `json:"end_time"`
    statusCode      int        `json:"status_code"`
    timestamp time.Time     `json:"timestamp"` // When a tick occur
    latency   time.Duration `json:"latency"`   // Round Trip Latency
    sentBytes  uint64        `json:"sent_bytes"`
    receivedBytes   uint64        `json:"received_bytes"`
    error     string        `json:"error"`
}
