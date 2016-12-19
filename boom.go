package main

import (
    "log"
    "regexp"
    "strings"
    "net/http"
    "bufio"
    "os"
)

var httpMethodChecker = regexp.MustCompile("^(HEAD|GET|PUT|POST|PATCH|OPTIONS|DELETE) ")

const headerSplitChar = ";"
const cookieSplitChar = headerSplitChar
const bodyReadFilePrefix = "@@"

func boom(boomOpts *BoomOptions) {

    targetOpts := parseTargetOptions(boomOpts)

    target, _ := newTargetWithOptions(targetOpts)

    missileOpts := parseMissileOptions(boomOpts)

    missile, _ := newMissile(missileOpts)

    harmsChan := missile.launch(target, boomOpts.requestPerSec, boomOpts.requestDuration)

    for hc := range harmsChan {
        log.Printf("Code: %d, Timestamp:%s, Latency:%s", hc.code, hc.timestamp, hc.latency)
    }
}

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

    /*if boomOpts.localAddress != "" {
        addr, err := net.ResolveIPAddr("ip", boomOpts.localAddress)
        if err != nil {
            log.Fatalf("Parse localAddress: %s error %s", boomOpts.localAddress, err.Error())
        }
        missileOpts.localAddr = addr
    } else {
        missileOpts.localAddr = net.IPAddr{IP: net.IPv4zero}
    }*/
    missileOpts.keepAlive = missileOpts.keepAlive

    // TODO support tlsConfig and http2Enable and maxRedirects
    return missileOpts

}

func checkHttpMethod(t string) bool {

    return httpMethodChecker.MatchString(t)
}
