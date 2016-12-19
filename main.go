package main

import (
    "flag"
    "runtime"
    "os"
    "fmt"
    "time"
    "log"
)

// Version of boom
const BoomVersion = "1.0"

var (
    // -cpu: The cpu to use
    cpuToUse int
    // -help: Show help message then exit
    showHelpUsage bool
    // -l: Enable log output
    showLogs bool
    // -V: Show version of boom then exit
    showVersion bool
)

// Options of boom
type BoomOptions struct {
    // -A: Supply BASIC Authentication credentials to the server.
    // The username and password are separated by a single : .
    authentication             string

    // -C: Add a Cookie: line to the request like: cookie-name=value
    requestCookies             string

    // -c: Content-type header to use for POST/PUT data,
    // eg. application/x-www-form-urlencoded. Default is text/plain.
    requestPostDataContentType string

    // -D: File or just a string containing data to POST. Remember to also set -c.
    requestPostData            string

    // -g: Number of threads(goroutines) to perform for the test.
    requestGoroutines          int

    // -H: Append extra headers to the request like: head-type:value
    requestHeaders             string

    // -k: Enable the HTTP KeepAlive feature
    enableKeepAlive            bool

    // -la: local address
    localAddress               string

    // -m: Custom HTTP method for the requests.
    requestMethod              string

    // -t: Duration of this test.
    requestDuration            time.Duration

    // -u： The url to request
    requestUrl                 string
    // -o: Output the reports in specified location
    resultOutput               string

    // -R: Generate reports in [text, json, plot]
    generateReports            string
    // -r: Number of requests to perform at one sec.
    requestPerSec              int
    // -P: Specify SSL/TLS protocol .
    requestProtocol            string
    // -s: Maximum number of seconds to wait before a request times out.
    requestTimeout             time.Duration
}


// Parse command line args
func parseArgs() (*BoomOptions, error) {
    boomOpts := &BoomOptions{
        requestGoroutines:1000,
        requestDuration:1,
        enableKeepAlive:false,
        requestTimeout:30,
    }
    flag.StringVar(&boomOpts.authentication, "A", "", "Supply BASIC Authentication credentials to the server. " +
        "The username and password are separated by a single : .")
    flag.StringVar(&boomOpts.requestCookies, "C", "", "Add a Cookie: line to the request like: cookie-name=value")
    flag.IntVar(&cpuToUse, "cpu", 1, "The cpu to use when sending requests")
    flag.StringVar(&boomOpts.requestPostDataContentType, "c", "", "Content-type header to use for POST/PUT data, " +
        "eg. application/x-www-form-urlencoded. Default is text/plain.")
    flag.StringVar(&boomOpts.requestPostData, "D", "", "File or just a string containing data to POST. Remember to also set -c.")
    flag.IntVar(&boomOpts.requestGoroutines, "g", 100, " Number of threads(goroutines) to perform for the test.")
    flag.StringVar(&boomOpts.requestHeaders, "H", "", "Append extra headers to the request like: head-type:value")
    flag.BoolVar(&boomOpts.enableKeepAlive, "k", false, "Enable the HTTP KeepAlive feature")
    flag.BoolVar(&showLogs, "l", false, "Enable log output")
    flag.StringVar(&boomOpts.localAddress, "la", "", "Local address")
    flag.StringVar(&boomOpts.requestMethod, "m", "GET", "Custom HTTP method for the requests.")
    flag.DurationVar(&boomOpts.requestDuration, "t", time.Second, "Duration of this test.")
    flag.StringVar(&boomOpts.requestUrl, "u", "", "The url to request")
    flag.StringVar(&boomOpts.resultOutput, "o", "", "Output the reports in specified location")
    flag.StringVar(&boomOpts.requestProtocol, "P", "HTTP", "Specify SSL/TLS protocol .")
    flag.DurationVar(&boomOpts.requestTimeout, "s", 30 * time.Second, "Maximum number of seconds to wait before a request times out.")
    flag.IntVar(&boomOpts.requestPerSec, "r", 50, "Number of requests to perform at one sec.")
    flag.StringVar(&boomOpts.generateReports, "R", "", "Generate reports in [text, json, plot]")
    flag.BoolVar(&showVersion, "V", false, " Show version of boom then exit")
    flag.Parse()
    return boomOpts, nil
}


// Show usage information
func usage() {
    usage := `aaaa`
    fmt.Print(usage)
    os.Exit(0)
}

// Main entrance
func main() {
    boomOpts, err := parseArgs()
    if err != nil {
        log.Fatal("Error accured when parsing command line args:" + err.Error())
        os.Exit(1)
    }
    if showVersion {
        fmt.Println(BoomVersion)
        os.Exit(0)
    }
    if showHelpUsage {
        usage()
        os.Exit(0)
    }
    // set GOMAXPROCS
    runtime.GOMAXPROCS(cpuToUse)

    // start boom
    boom(boomOpts)
}
