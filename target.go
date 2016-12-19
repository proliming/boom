package main

import (
    "net/http"
    "bytes"
)

type TargetOptions struct {
    method string
    url    string
    body   []byte
    header http.Header
    cookie http.Cookie
}

// target 代表一个请求的目标
// 多个missile可以向同一个target发起攻击(请求)
type Target struct {
    method string
    url    string
    body   []byte
    header http.Header
    cookie http.Cookie
}


func newTargetWithUrl(url string) (target *Target) {
    target = &Target{
        url:url,
        method:"GET",
        body:[]byte{},
        header:make(map[string][]string),
        cookie:http.Cookie{},
    }
    return target
}

func newTargetWithOptions(targetOpts *TargetOptions) (target *Target, err error) {
    target = &Target{
        url:targetOpts.url,
        method:targetOpts.method,
        body: targetOpts.body,
        header:targetOpts.header,
        cookie:targetOpts.cookie,
    }
    return target, nil
}


// 返回一个 *http.Request 的封装
func (t *Target) request() (*http.Request, error) {
    req, err := http.NewRequest(t.method, t.url, bytes.NewReader(t.body))
    if err != nil {
        return nil, err
    }
    for k, vs := range t.header {
        req.Header[k] = make([]string, len(vs))
        copy(req.Header[k], vs)
    }
    if host := req.Header.Get("Host"); host != "" {
        req.Host = host
    }
    return req, nil
}