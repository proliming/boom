package main

import (
    "fmt"
    "time"
    "sort"
    "log"
)

type ServerInfo struct {
    URL      string
    Software string
    HostName string
}

//
type Report struct {
    ServerInfo         *ServerInfo `json:"server_info"`
    ConcurrencyLevel   int `json:"concurrency_level"`
    TimeTaken          float64 `json:"time_taken"`
    CompletedRequests  int `json:"completed_requests"`
    FailedRequests     int `json:"failed_requests"`
    TotalSendBytes     uint64 `json:"total_send_bytes"`
    TotalReceivedBytes uint64 `json:"total_received_bytes"`
    TotalTransferred   uint64 `json:"total_transfered"`
    RequestPerSecond   float64 `json:"request_per_second"`
    TimePerRequest     float64 `json:"time_per_request"`
    TransferRate       float64 `json:"transfer_rate"`
}

type ReportTimeSlice [] time.Time
type ReportLatencySlice []time.Duration

func generateReport(harmChan <-chan *Harm, boomOpts *BoomOptions) (report *Report) {
    var (
        harmCount = 0
        failedRequests = 0
        completedRequests = 0
        totalSendBytes = uint64(0)
        totalReceivedBytes = uint64(0)
        totalLatency = float64(0)
        timestamps = make([]time.Time, 0)
        latencies = make([]time.Duration, 0)
    )
    report = &Report{}
    report.ConcurrencyLevel = boomOpts.requestGoroutines

    //var firstResultComeTime time.Time
    var lastResultComeTime time.Time

    for harm := range harmChan {
        lastResultComeTime = time.Now()
        harmCount++
        completedRequests++
        if harmCount % 100 == 0 {
            fmt.Printf("Completed %d requests.\n", harmCount)
        }
        if harm.error != "" {
            failedRequests++
        }
        totalSendBytes += harm.sentBytes
        totalReceivedBytes += harm.receivedBytes
        totalLatency += harm.latency.Seconds()
        timestamps = append(timestamps, harm.timestamp)
        latencies = append(latencies, harm.latency)

    }

    report.CompletedRequests = completedRequests
    report.FailedRequests = failedRequests
    report.TotalReceivedBytes = totalReceivedBytes
    report.TotalSendBytes = totalSendBytes
    report.TotalTransferred = totalSendBytes + totalReceivedBytes

    sort.Sort(ReportTimeSlice(timestamps))
    sort.Sort(ReportLatencySlice(latencies))

    startTime := timestamps[0]
    //endTime := timestamps[len(timestamps) - 1]
    duration := lastResultComeTime.Sub(startTime)
    report.TimeTaken = duration.Seconds()
    report.RequestPerSecond = float64(completedRequests) / duration.Seconds()
    report.TimePerRequest = totalLatency / float64(completedRequests)
    /*if boomOpts.totalRequests > 0 {
        report.RequestPerSecond = float64(completedRequests) / duration.Seconds()
        report.TimePerRequest = totalLatency / float64(completedRequests)
    } else {
        report.RequestPerSecond = float64(boomOpts.requestPerSec)
        report.TimePerRequest = float64(1) / float64(boomOpts.requestPerSec) / 1000
    }*/

    report.TransferRate = float64(report.TotalTransferred) / duration.Seconds()

    log.Println("Report generated done!")

    log.Println("Generating reports...")
    if boomOpts.resultOutput != "Stdout" {
        report.writeToFile(boomOpts.resultOutput)
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
    fmt.Printf("Total send: %d bytes\n", r.TotalSendBytes)
    fmt.Printf("Total received: %d bytes\n", r.TotalReceivedBytes)
    fmt.Printf("Total transferred: %d bytes\n", r.TotalTransferred)
    fmt.Printf("Requests per second: %.3fs (mean)\n", r.RequestPerSecond)
    fmt.Printf("Time per request: %.6fs (mean)\n", r.TimePerRequest)
    fmt.Printf("Transfer rate: %.3f bytes/s (mean)\n", r.TransferRate)

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

