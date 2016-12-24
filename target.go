package main

import (
    "regexp"
    "net/http"
    "strings"
    "time"
    "bytes"
)

const (
    defaultMethod = "GET"
)

// A target is just a URL with some extra properties.
type Target struct {
    method string
    Url    string
    Body   []byte
    header http.Header
    cookie http.Cookie
}

// Use to check http methods
var httpMethodChecker = regexp.MustCompile("^(HEAD|GET|PUT|POST|PATCH|OPTIONS|DELETE)")


// Create a target with the specified url.
func NewTarget(url string) (t *Target) {
    t = &Target{Url:url, method:defaultMethod, header:http.Header{}}
    return t
}

// Set the request method
func (t *Target) SetMethod(m string) error {
    if httpMethodChecker.MatchString(m) {
        t.method = m
    } else {
        t.method = defaultMethod
        return errInvalidHttpMethod
    }
    return nil
}

// Add adds the key, value pair to the header.
// It appends to any existing values associated with key.
func (t *Target) AddHeader(key, value string) {
    t.header.Add(key, value)
}

func (t *Target) SetCookie(key string, value interface{}) {
    switch strings.ToLower(key) {
    case "name":
        t.cookie.Name = value.(string)
    case "value":
        t.cookie.Value = value.(string)
    case "path":
        t.cookie.Path = value.(string)
    case "domain":
        t.cookie.Domain = value.(string)
    case "expires":
        exp, _ := time.Parse("ANSIC", value.(string))
        t.cookie.Expires = exp
    case "maxage":
        t.cookie.MaxAge = value.(int)
    default:
        panic("Not supported for now. key:" + key + ", value:" + value.(string))
    }
}

func (t *Target) Request() (*http.Request, error) {
    req, err := http.NewRequest(t.method, t.Url, bytes.NewReader(t.Body))
    if err != nil {
        return nil, err
    }
    for k, vs := range t.header {
        req.Header[k] = make([]string, len(vs))
        copy(req.Header[k], vs)
    }
    req.AddCookie(&t.cookie)
    if host := req.Header.Get("Host"); host != "" {
        req.Host = host
    }
    return req, nil
}