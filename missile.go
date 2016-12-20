package main

import (
    "net"
    "net/http"
    "time"
    "sync"
    "io"
    "io/ioutil"
    "crypto/tls"
    "golang.org/x/net/http2"
    "log"
)
// Missile is a wrapper of http.Client and some properties
type Missile struct {
    dialer     *net.Dialer
    client     http.Client
    stopAttack chan struct{}
    launchers  int
    redirects  int
}

// Options of a Missile
type MissileOptions struct {
    timeout            time.Duration
    launchers          int
    maxIdleConnections int
    keepAlive          bool
    http2Enable        bool
    maxRedirects       int
    localAddr          *net.IPAddr
    tlsConfig          *tls.Config
}

const (
    defaultTimeout = 30 * time.Second
    defaultConnections = 10000
    defaultLaunchers = 100
    noFollow = -1
)

var (
    defaultLocalAddr = net.IPAddr{IP: net.IPv4zero}
    defaultTLSConfig = &tls.Config{InsecureSkipVerify: true}
)

// Create a missile with the given options
func newMissile(missileOptions *MissileOptions) *Missile {

    missile := &Missile{stopAttack: make(chan struct{}), launchers: defaultLaunchers}

    missile.dialer = &net.Dialer{
        LocalAddr: &net.TCPAddr{IP: defaultLocalAddr.IP, Zone: defaultLocalAddr.Zone},
        KeepAlive: 30 * time.Second,
        Timeout:   defaultTimeout,
    }

    missile.client = http.Client{
        Transport: &http.Transport{
            Proxy: http.ProxyFromEnvironment,
            Dial:  missile.dialer.Dial,
            ResponseHeaderTimeout: defaultTimeout,
            TLSClientConfig:       defaultTLSConfig,
            TLSHandshakeTimeout:   10 * time.Second,
            MaxIdleConnsPerHost:   defaultConnections,
        },
    }

    if missileOptions != nil {
        tr := missile.client.Transport.(*http.Transport)

        missile.launchers = missileOptions.launchers;
        missile.dialer.Timeout = missileOptions.timeout
        tr.ResponseHeaderTimeout = missileOptions.timeout
        tr.MaxIdleConnsPerHost = missileOptions.maxIdleConnections
        if missileOptions.tlsConfig != nil {
            tr.TLSClientConfig = missileOptions.tlsConfig
        }
        if !missileOptions.keepAlive {
            tr.DisableKeepAlives = true
            missile.dialer.KeepAlive = 0
        }
        if missileOptions.http2Enable {
            http2.ConfigureTransport(tr)
        } else {
            tr.TLSNextProto = map[string]func(string, *tls.Conn) http.RoundTripper{}
        }
    }
    return missile
}


// Launch the Missile
func (missile *Missile) launch(target *Target, totalHits int, attackPerSec int, du time.Duration) <-chan *Harm {
    var launchers sync.WaitGroup
    harms := make(chan *Harm)
    ticks := make(chan time.Time)
    for i := 0; i < missile.launchers; i++ {
        launchers.Add(1)
        go missile.fire(target, &launchers, ticks, harms)
    }
    // Start an
    go func() {
        defer close(harms)
        defer launchers.Wait()
        defer close(ticks)
        if totalHits > 0 {
            done := 0
            for {
                //time.Sleep(1 * time.Second)
                select {
                case ticks <- time.Now():
                    if done++; done == totalHits {
                        return
                    }
                case <-missile.stopAttack:
                    return
                default:
                // all workers are blocked. start one more and try again
                    launchers.Add(1)
                    go missile.fire(target, &launchers, ticks, harms)
                }
            }
        } else {
            interval := 1e9 / attackPerSec
            hits := attackPerSec * int(du.Seconds())
            began, done := time.Now(), 0
            for {
                now, next := time.Now(), began.Add(time.Duration(done * interval))
                time.Sleep(next.Sub(now))
                select {
                case ticks <- max(next, now):
                    if done++; done == hits {
                        return
                    }
                case <-missile.stopAttack:
                    return
                default:
                // all workers are blocked. start one more and try again
                    launchers.Add(1)
                    go missile.fire(target, &launchers, ticks, harms)
                }
            }
        }
    }()
    return harms
}

// Fire
func (missile *Missile) fire(target *Target, launchers *sync.WaitGroup, ticks <-chan time.Time, results chan <-*Harm) {
    defer launchers.Done()
    for tk := range ticks {
        results <- missile.hit(target, tk)
    }
}

// Hit the Target
func (missile *Missile) hit(target *Target, hitStartTime time.Time) *Harm {

    hitResult := Harm{timestamp: hitStartTime}
    var hitError error

    defer func() {
        hitResult.latency = time.Since(hitStartTime)
        if hitError != nil {
            hitResult.error = hitError.Error()
        }
    }()

    req, err := target.request()
    if err != nil {
        return &hitResult
    }

    resp, err := missile.client.Do(req)
    if err != nil {
        return &hitResult
    }
    defer resp.Body.Close()

    in, err := io.Copy(ioutil.Discard, resp.Body)
    if err != nil {
        return &hitResult
    }

    hitResult.bytesIn = uint64(in)

    if req.ContentLength != -1 {
        hitResult.bytesOut = uint64(req.ContentLength)
    }

    if hitResult.code = resp.StatusCode; hitResult.code < 200 || hitResult.code >= 400 {
        hitResult.error = resp.Status
    }
    return &hitResult
}

// Stop stops the current attack.
func (missile *Missile) stop(boomOpts * BoomOptions) {
    log.Println("Missle will stop.")
    select {
    case <-missile.stopAttack:
        return
    default:
        close(missile.stopAttack)
    }

}

func max(a, b time.Time) time.Time {
    if a.After(b) {
        return a
    }
    return b
}
