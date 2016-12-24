package main

import (
    "flag"
    "runtime"
    "os"
    "fmt"
    "time"
    "log"
    "io/ioutil"
)

// Version of boom
const BoomVersion = "Boom version 1.2"

// Use to calculate the total time cost when running this test.
var boomStartTime time.Time;

var (
    // -cpu: The cpu to use
    cpuToUse int
    // -l: Enable log output
    showLogs bool
    // -V: Show version of boom then exit
    showVersion bool
)


// Parse command line args
func parseArgs() *BoomOptions {
    boomOpts := &BoomOptions{
        RequestMethod:"GET",
        RequestPerSec:1000,
        RequestGoroutines:100,
        RequestDuration: 1 * time.Second,
        EnableKeepAlive:false,
        RequestTimeout: 30 * time.Second,
    }
    flag.StringVar(&boomOpts.Authentication, "A", "", "Supply BASIC Authentication credentials to the server. " +
        "The username and password are separated by a single : .")
    flag.StringVar(&boomOpts.RequestCookies, "C", "", "Add a Cookie: line to the request like: cookie-name=value")
    flag.IntVar(&cpuToUse, "cpu", 1, "The cpu to use when sending requests")
    flag.StringVar(&boomOpts.RequestPostDataContentType, "c", "", "Content-type header to use for POST/PUT data, " +
        "eg. application/x-www-form-urlencoded. Default is text/plain.")
    flag.StringVar(&boomOpts.RequestPostData, "D", "", "File or just a string containing data to POST. Remember to " +
        "also set -c." + "When using a file for input, remember add '@@' prefix to the file path. eg. @@/home/work/a.json")
    flag.IntVar(&boomOpts.RequestGoroutines, "g", 100, " Number of threads(goroutines) to perform for the test.")
    flag.StringVar(&boomOpts.RequestHeaders, "H", "", "Append extra headers to the request like: head-type:value")
    flag.BoolVar(&boomOpts.EnableKeepAlive, "k", false, "Enable the HTTP KeepAlive feature")
    flag.BoolVar(&showLogs, "l", false, "Enable log output")
    flag.StringVar(&boomOpts.LocalAddr, "la", "", "Local address  to bind to when making outgoing connections.")
    flag.StringVar(&boomOpts.RequestMethod, "m", "GET", "Custom HTTP method for the requests.")
    flag.IntVar(&boomOpts.TotalRequests, "n", 0, "Number of requests to perform for the test. If this flag > 0, the " +
        "-t and -r will be ignore.")
    flag.DurationVar(&boomOpts.RequestDuration, "t", time.Second, "Duration of this test.")
    flag.StringVar(&boomOpts.URL, "u", "", "The url to request")
    flag.StringVar(&boomOpts.ResultOutput, "o", "Stdout", "Output the reports in specified location")
    flag.DurationVar(&boomOpts.RequestTimeout, "s", 30 * time.Second, "Maximum number of seconds to wait before a " +
        "request times out.")
    flag.IntVar(&boomOpts.RequestPerSec, "r", 50, "Number of requests to perform at one sec.")
    flag.BoolVar(&showVersion, "V", false, " Show version of boom then exit")
    flag.Parse()

    return boomOpts
}


// Show usage information
func usage() {
    usage := ``
    fmt.Print(usage)
    os.Exit(0)
}

func welcome() {

    welcome := `
This is Boom, Version 1.2
Copyright (c) 2016- Li Ming, http://www.waymou.com/
Licensed to The Apache Software Foundation, http://www.apache.org/

This test will take some time. Please wait for a while :-)
`
    fmt.Println(welcome)

}

// Main entrance
func main() {
    boomOpts := parseArgs()

    if showVersion {
        fmt.Println(BoomVersion)
        os.Exit(0)
    }
    if len(os.Args[1:]) == 0 {
        usage()
        os.Exit(0)
    }
    if !showLogs {
        log.SetOutput(ioutil.Discard)
    } else {
        log.SetOutput(os.Stdout)
    }

    // set GOMAXPROCS
    runtime.GOMAXPROCS(cpuToUse)

    log.Printf("Starting boom ...")
    // start boom
    boomStartTime = time.Now()
    welcome()
    Boom(boomOpts)
}
