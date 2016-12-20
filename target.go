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

// Target is a wrapper of http.Request
type Target struct {
    method string
    url    string
    body   []byte
    header http.Header
    cookie http.Cookie
}

// Create a Target with options.
func newTarget(targetOpts *TargetOptions) (target *Target) {
    target = &Target{
        url:targetOpts.url,
        method:targetOpts.method,
        body: targetOpts.body,
        header:targetOpts.header,
        cookie:targetOpts.cookie,
    }
    return target
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