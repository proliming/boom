package main

import "time"


// Standards for the result by launching a Missile
type Damage struct {
    StartTime     time.Time `json:"start_time"`
    EndTime       time.Time `json:"end_time"`
    StatusCode    int        `json:"status_code"`
    Timestamp     time.Time     `json:"timestamp"` // When a tick occur
    Latency       time.Duration `json:"latency"`   // Round Trip Latency
    SentBytes     uint64        `json:"sent_bytes"`
    ReceivedBytes uint64        `json:"received_bytes"`
    Error         string        `json:"error"`
}
