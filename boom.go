package main

import (
    "log"
    "regexp"
    "strings"
    "net/http"
    "bufio"
    "os"
)

// Use to check http methods
var httpMethodChecker = regexp.MustCompile("^(HEAD|GET|PUT|POST|PATCH|OPTIONS|DELETE) ")

const headerSplitChar = ";"
const cookieSplitChar = headerSplitChar
const bodyReadFilePrefix = "@@"

// This method will create Target and Missile
// then launch the Missile
func boom(boomOpts *BoomOptions) {

    targetOpts := parseTargetOptions(boomOpts)
    missileOpts := parseMissileOptions(boomOpts)

    // now create target and missile
    target := newTargetWithOptions(targetOpts)
    missile := newMissile(missileOpts)

    // launch
    harmsResult := missile.launch(target, boomOpts.totalRequests, boomOpts.requestPerSec, boomOpts.requestDuration)

    log.Println("The missile launched!")
    // collects the report
    reports := generateReport(harmsResult, boomOpts)

    log.Println("Generating reports...")
    if boomOpts.resultOutput != "Stdout" {
        reports.writeToFile(boomOpts.resultOutput)
    } else {
        reports.prettyPrintToConsole()
    }
}

// Parse the TargetOptions
func parseTargetOptions(boomOpts *BoomOptions) (targetOpts *TargetOptions) {
    log.Println("Parsing target options...")
    targetOpts = &TargetOptions{}
    method := boomOpts.requestMethod

    if checkHttpMethod(method) {
        targetOpts.method = method
    } else {
        targetOpts.method = "GET"
    }
    // TODO url check?
    if len(boomOpts.requestUrl) > 0 {
        targetOpts.url = boomOpts.requestUrl
    } else {
        log.Fatal("Url not set for boom.")
    }

    httpHeader := http.Header{}
    if boomOpts.requestHeaders != "" {
        // header
        headerTokens := strings.Split(boomOpts.requestHeaders, headerSplitChar)
        for _, sh := range headerTokens {
            headerValue := strings.Split(sh, ":")
            if len(headerValue) != 2 {
                log.Fatalf("Not valid http header:%s", sh)
            }
            httpHeader.Add(strings.TrimSpace(headerValue[0]), strings.TrimSpace(headerValue[1]))
        }
        targetOpts.header = httpHeader
    }

    //TODO for now cookie not supported

    // post data
    if boomOpts.requestPostData != "" {
        body := boomOpts.requestPostData
        if strings.HasPrefix(body, bodyReadFilePrefix) {
            bodyContentFile := strings.TrimPrefix(body, bodyReadFilePrefix)
            f, err := os.Open(bodyContentFile)
            if err != nil {
                log.Fatalf("Can't open file specified in :%s", bodyContentFile)
            }
            reader := bufio.NewReader(f)
            bodyBytes := []byte{}
            n, err := reader.Read(bodyBytes)
            if n > 0 && err == nil {
                targetOpts.body = bodyBytes
            } else {
                log.Fatalf("Read file to post error :%s", bodyContentFile)
            }
        } else {
            targetOpts.body = []byte(strings.TrimSpace(body))
        }
        if boomOpts.requestPostDataContentType == "" {
            log.Println("Post data setted but no Content-Type. Will using text/plain")
            httpHeader.Add("Content-Type", "text/plain")
        } else {
            httpHeader.Add("Content-Type", strings.TrimSpace(boomOpts.requestPostDataContentType))
        }
    }

    return targetOpts

}

// Parse the MissileOptions
func parseMissileOptions(boomOpts *BoomOptions) (missileOpts *MissileOptions) {
    log.Println("Parsing missile options...")
    missileOpts = &MissileOptions{}
    if boomOpts.requestTimeout > 0 {
        missileOpts.timeout = boomOpts.requestTimeout
    } else {
        missileOpts.timeout = defaultTimeout
    }
    if boomOpts.requestGoroutines > 0 {
        missileOpts.launchers = boomOpts.requestGoroutines
    } else {
        missileOpts.launchers = defaultLaunchers
    }
    // TODO support custom setting
    missileOpts.maxIdleConnections = defaultConnections

    missileOpts.keepAlive = missileOpts.keepAlive

    log.Printf("Missile launchers:%d, timeout:%t, keepAlive:%b", missileOpts.launchers, missileOpts.timeout,
        missileOpts.keepAlive)
    // TODO support tlsConfig and http2Enable and maxRedirects
    return missileOpts

}

// check
func checkHttpMethod(m string) bool {
    return httpMethodChecker.MatchString(m)
}
