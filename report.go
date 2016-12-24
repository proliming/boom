package main

import (
    "fmt"
    "time"
    "sort"
    "log"
)

// Server(Target) information
type ServerInfo struct {
    URL      string
    Software string
    HostName string
    Port     int
}

//
type Report struct {
    ServerInfo                *ServerInfo `json:"server_info"`
    ConcurrencyLevel          int `json:"concurrency_level"` // how many goroutines in this test
    TimeTaken                 float64 `json:"time_taken"`
    CompletedRequests         int `json:"completed_requests"`
    FailedRequests            int `json:"failed_requests"`
    SuccessRate               float64 `json:"success_rate"`
    TotalSentBytes            uint64 `json:"total_send_bytes"`
    TotalReceivedBytes        uint64 `json:"total_received_bytes"`
    TotalTransferred          uint64 `json:"total_transfered"`
    RequestPerSecond          float64 `json:"request_per_second"`
    TimePerRequest            float64 `json:"time_per_request"`
    TimePerRequestConcurrency float64 `json:"time_per_request_concurrency"`
    TransferRate              float64 `json:"transfer_rate"`
    MinLatency                float64 `json:"min_latency"`
    MaxLatency                float64 `json:"max_latency"`
    MeanLatency               float64 `json:"mean_latency"`
}

type ReportTimeSlice []time.Time
type ReportLatencySlice []time.Duration

// Damages buffer use to store the damages received from damage channel.
// Boom generates the final report when the channel is closed or the process is killed.
var damagesBuffer = make([]*Damage, 0)


// Receive the damages come from channel
func collectDamage(damage *Damage) {
    damagesBuffer = append(damagesBuffer, damage)
}
func receiveDamages(damageCh <-chan *Damage) {
    for damage := range damageCh {
        damagesBuffer = append(damagesBuffer, damage)
    }
}

// Create a Report
func createReport(boomOpts *BoomOptions) (report *Report) {
    // fmt.Println("Generating boom report, please be patient... :-) ")
    if len(damagesBuffer) <= 0 {
        log.Println("No damages.")
        return nil
    }
    var (
        failedRequests = 0
        completedRequests = 0
        totalSentBytes = uint64(0)
        totalReceivedBytes = uint64(0)
        totalLatency = float64(0)
        startTimestamps = make(ReportTimeSlice, 0)
        latencies = make(ReportLatencySlice, 0)
        endTimeStamps = make(ReportTimeSlice, 0)
    )
    for _, damage := range damagesBuffer {
        completedRequests++
        if damage.Error != "" {
            log.Println(damage.Error)
            failedRequests++
        }
        totalSentBytes += damage.SentBytes
        totalReceivedBytes += damage.ReceivedBytes
        startTimestamps = append(startTimestamps, damage.Timestamp)
        latencies = append(latencies, damage.Latency)
        endTimeStamps = append(endTimeStamps, damage.EndTime)
        totalLatency += damage.Latency.Seconds()
    }

    sort.Sort(startTimestamps)
    sort.Sort(endTimeStamps)
    sort.Sort(latencies)

    report = &Report{}

    report.ConcurrencyLevel = boomOpts.RequestGoroutines

    // requests
    report.CompletedRequests = completedRequests
    report.FailedRequests = failedRequests
    report.SuccessRate = float64(completedRequests - failedRequests) / float64(completedRequests)

    // bytes
    report.TotalReceivedBytes = totalReceivedBytes
    report.TotalSentBytes = totalSentBytes
    report.TotalTransferred = totalSentBytes + totalReceivedBytes

    firstRequestSendTime := startTimestamps[0]
    lastRequestCompletedTime := endTimeStamps[len(endTimeStamps) - 1]
    lastRequestCompletedTime.Sub(firstRequestSendTime)

    // TimeTaken = (First request sent) - (Last request response)
    report.TimeTaken = lastRequestCompletedTime.Sub(firstRequestSendTime).Seconds()
    // RequestPerSecond = (Complete requests) / (Time taken for tests)
    report.RequestPerSecond = float64(completedRequests) / report.TimeTaken
    // TransferRate = (Total transferred bytes) / (Time taken for tests)
    report.TransferRate = float64(report.TotalTransferred) / report.TimeTaken

    // TimePerRequest = Time taken for tests /（ Complete requests / Concurrency Level）
    report.TimePerRequest = report.TimeTaken / (float64(completedRequests) / float64(report.ConcurrencyLevel))
    // TimeRerRequestConcurrency: across all concurrent requests
    report.TimePerRequestConcurrency = report.TimeTaken / float64(completedRequests)

    // Latency
    report.MeanLatency = totalLatency / float64(completedRequests)
    report.MaxLatency = latencies[len(latencies) - 1].Seconds()
    report.MinLatency = latencies[0].Seconds()

    if boomOpts.ResultOutput != "Stdout" {
        report.writeToFile(boomOpts.ResultOutput)
    } else {
        report.prettyPrintToConsole()
    }
    return report

}

// Print report content to console
func (r *Report) prettyPrintToConsole() {

    fmt.Printf("Concurrency Level: %d\n", r.ConcurrencyLevel)
    fmt.Printf("Time taken for tests: %.6fs \n", r.TimeTaken)

    fmt.Printf("Complete requests: %d\n", r.CompletedRequests)
    fmt.Printf("Failed requests: %d\n", r.FailedRequests)
    fmt.Printf("Success Rate: %.2f %% \n", r.SuccessRate * 100)

    fmt.Printf("Total sent: %d bytes\n", r.TotalSentBytes)
    fmt.Printf("Total received: %d bytes\n", r.TotalReceivedBytes)
    fmt.Printf("Total transferred: %d bytes\n", r.TotalTransferred)
    fmt.Printf("Transfer rate: %.3f bytes/s (mean)\n", r.TransferRate)

    fmt.Printf("Requests per second: %.3f (mean)\n", r.RequestPerSecond)
    fmt.Printf("Time per request: %.3fms (mean)\n", r.TimePerRequest * 1000)
    fmt.Printf("Time per request concurrency: %.3fms (mean)\n", r.TimePerRequestConcurrency * 1000)
    fmt.Printf("Latency(min,mean,max): %.3fms, %.3fms ,%.3fms \n", r.MinLatency * 1000, r.MeanLatency * 1000, r.MaxLatency * 1000)

}

// Write report content to file
func (r *Report) writeToFile(file string) {

}

// Forward request for length
func (p ReportTimeSlice) Len() int {
    return len(p)
}

// Define compare
func (p ReportTimeSlice) Less(i, j int) bool {
    return p[i].Before(p[j])
}

// Define swap over an array
func (p ReportTimeSlice) Swap(i, j int) {
    p[i], p[j] = p[j], p[i]
}
// Forward request for length
func (p ReportLatencySlice) Len() int {
    return len(p)
}

// Define compare
func (p ReportLatencySlice) Less(i, j int) bool {
    return p[i].Nanoseconds() < p[j].Nanoseconds()
}

// Define swap over an array
func (p ReportLatencySlice) Swap(i, j int) {
    p[i], p[j] = p[j], p[i]
}

