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
    target, _ := newTargetWithOptions(targetOpts)
    missile, _ := newMissile(missileOpts)

    // launch
    harmsResult := missile.launch(target, boomOpts.requestPerSec, boomOpts.requestDuration)

    // collects the report
    reports, _ := generateReport(harmsResult)

    if boomOpts.resultOutput != "Stdout" {
        reports.writeToFile(boomOpts.resultOutput)
    } else {
        reports.prettyPrintToConsole()
    }
}

// Parse the TargetOptions
func parseTargetOptions(boomOpts *BoomOptions) (targetOpts *TargetOptions) {
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
    }

    if boomOpts.requestHeaders != "" {
        // header
        httpHeader := http.Header{}
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

    // for now cookie not supported
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
            }
        } else {
            targetOpts.body = []byte(strings.TrimSpace(body))
        }
    }

    return targetOpts

}

// Parse the MissileOptions
func parseMissileOptions(boomOpts *BoomOptions) (missileOpts *MissileOptions) {
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

    // TODO support tlsConfig and http2Enable and maxRedirects
    return missileOpts

}

// check
func checkHttpMethod(m string) bool {
    return httpMethodChecker.MatchString(m)
}
