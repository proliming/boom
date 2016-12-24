package main

import (
    "time"
    "log"
    "strings"
    "os"
    "bufio"
    "os/signal"
)

// Options of boom
type BoomOptions struct {
    // -A: Supply BASIC Authentication credentials to the server.
    // The username and password are separated by a single : .
    Authentication             string

    // -a: Local address
    LocalAddr                  string

    // -C: Add a Cookie: line to the request like: cookie-name=value
    RequestCookies             string

    // -c: Content-type header to use for POST/PUT data,
    // eg. application/x-www-form-urlencoded. Default is text/plain.
    RequestPostDataContentType string

    // -D: File or just a string containing data to POST. Remember to also set -c.
    RequestPostData            string

    // -g: Number of threads(goroutines) to perform for the test.
    RequestGoroutines          int

    // -H: Append extra headers to the request like: head-type:value
    RequestHeaders             string

    // -k: Enable the HTTP KeepAlive feature
    EnableKeepAlive            bool

    // -m: Custom HTTP method for the requests.
    RequestMethod              string

    // -n: Number of requests to perform for the test. If this flag > 0, the -t and -r will be ignore.
    TotalRequests              int

    // -t: Duration of this test, Remember to set -r.
    RequestDuration            time.Duration

    // -u： The url to request
    URL                        string
    // -o: Output the reports in specified location
    ResultOutput               string

    // -r: Number of requests to perform at one sec.
    RequestPerSec              int

    // -s: Maximum number of seconds to wait before a request times out.
    RequestTimeout             time.Duration
}

func Boom(opts *BoomOptions) {
    err := checkOpts(opts)
    if err != nil {
        log.Fatal(err.Error())
    }
    target := createTarget(opts)
    log.Println("Target ready.")

    missile := createMissile(opts)
    log.Println("Missile ready.")



    damagesResult := missile.Launch(target, opts.TotalRequests, opts.RequestPerSec, opts.RequestDuration)

    log.Println("The missile launched!")

    killFlag := make(chan os.Signal, 1)
    signal.Notify(killFlag, os.Interrupt)

    for {
        select {
        case <-killFlag:
            missile.Stop()
            log.Println("Press CTRL+C")
            createReport(opts)
            return
        case r, ok := <-damagesResult:
            if !ok {
                createReport(opts)
                return
            } else {
                collectDamage(r)
            }
        }
    }

}

func createMissile(opts *BoomOptions) *Missile {

    cc := NewDefaultCtrlCenter()
    if opts.RequestTimeout > 0 {
        cc.Timeout = opts.RequestTimeout
    } else {
        cc.Timeout = defaultTimeout
    }
    if opts.RequestGoroutines > 0 {
        cc.Warheads = opts.RequestGoroutines
    } else {
        cc.Warheads = defaultWarheads
    }
    if opts.EnableKeepAlive {
        cc.KeepAlive = 30 * time.Second
    }
    return NewCustomMissile(cc)
}

func createTarget(opts *BoomOptions) *Target {
    target := NewTarget(opts.URL)

    target.SetMethod(opts.RequestMethod)

    if opts.RequestHeaders != "" {
        // header
        headerTokens := strings.Split(opts.RequestHeaders, ";")
        for _, sh := range headerTokens {
            headerValue := strings.Split(sh, ":")
            if len(headerValue) != 2 {
                log.Fatalf("Not valid http header:%s", sh)
            }
            target.AddHeader(strings.TrimSpace(headerValue[0]), strings.TrimSpace(headerValue[1]))
        }
    }

    // post data
    if opts.RequestPostData != "" {
        body := opts.RequestPostData
        if strings.HasPrefix(body, "@@") {
            bodyContentFile := strings.TrimPrefix(body, "@@")
            f, err := os.Open(bodyContentFile)
            if err != nil {
                log.Fatalf("Can't open file specified in :%s", bodyContentFile)
            }
            reader := bufio.NewReader(f)
            bodyBytes := []byte{}
            n, err := reader.Read(bodyBytes)
            if n > 0 && err == nil {
                target.Body = bodyBytes
            } else {
                log.Fatalf("Read file to post error :%s", bodyContentFile)
            }
        } else {
            target.Body = []byte(strings.TrimSpace(body))
        }
        if opts.RequestPostDataContentType == "" {
            log.Println("Post data setted but no Content-Type. Will using text/plain")
            target.AddHeader("Content-Type", "text/plain")
        } else {
            target.AddHeader("Content-Type", strings.TrimSpace(opts.RequestPostDataContentType))
        }
    }
    return target
}

func checkOpts(opts *BoomOptions) error {
    if opts == nil {
        return errNilBoomOpts
    }
    if opts.URL == "" {
        return errBoomOpts
    }
    // Some other check
    return nil
}

