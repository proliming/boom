package main

import "time"

// Standards for the result by launching a Missile
type Harm struct {
    code      int        `json:"code"`
    timestamp time.Time     `json:"timestamp"`
    latency   time.Duration `json:"latency"`
    bytesOut  uint64        `json:"bytes_out"`
    bytesIn   uint64        `json:"bytes_in"`
    error     string        `json:"error"`
}
