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
// A missile can carry many warheads means multi goroutines
type Missile struct {
    dialer     *net.Dialer
    client     http.Client
    stopAttack chan struct{}
    warheads   int
    redirects  int
}

// Options of a Missile
type MissileOptions struct {
    timeout            time.Duration
    warheads           int // How many warhead can this missile carry
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
    defaultWarheads = 100
    noFollow = -1
)

var (
    defaultLocalAddr = net.IPAddr{IP: net.IPv4zero}
    defaultTLSConfig = &tls.Config{InsecureSkipVerify: true}
)

// Create a missile with the given options
func newMissile(missileOptions *MissileOptions) *Missile {

    missile := &Missile{stopAttack: make(chan struct{}), warheads: defaultWarheads}

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

        missile.warheads = missileOptions.warheads;
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
func (missile *Missile) launch(target *Target, totalHits int, hitPerSecond int, du time.Duration) <-chan *Damage {
    var warheadsWaitGroup sync.WaitGroup
    damagesCh := make(chan *Damage)
    fireCmdCh := make(chan time.Time)

    // Each warhead standard for a single goroutine
    for i := 0; i < missile.warheads; i++ {
        warheadsWaitGroup.Add(1)
        go missile.fire(target, &warheadsWaitGroup, fireCmdCh, damagesCh)
    }
    go func() {
        defer close(damagesCh)
        defer warheadsWaitGroup.Wait()
        defer close(fireCmdCh)
        if totalHits > 0 {
            done := 0
            for {
                //time.Sleep(1 * time.Second)
                select {
                case fireCmdCh <- time.Now():
                    if done++; done == totalHits {
                        return
                    }
                case <-missile.stopAttack:
                    return
                default:
                // all warhead are blocked. start one more and try again
                    warheadsWaitGroup.Add(1)
                    go missile.fire(target, &warheadsWaitGroup, fireCmdCh, damagesCh)
                }
            }
        } else {
            //Interval non-negative nanosecond
            interval := 1e9 / hitPerSecond
            hitsSum := hitPerSecond * int(du.Seconds())
            began, done := time.Now(), 0
            for {
                now, next := time.Now(), began.Add(time.Duration(done * interval))
                time.Sleep(next.Sub(now))
                select {
                case fireCmdCh <- max(next, now):
                    if done++; done == hitsSum {
                        return
                    }
                case <-missile.stopAttack:
                    return
                default:
                // all workers are blocked. start one more and try again
                    warheadsWaitGroup.Add(1)
                    go missile.fire(target, &warheadsWaitGroup, fireCmdCh, damagesCh)
                }
            }
        }
    }()
    return damagesCh
}

// Fire
func (missile *Missile) fire(target *Target,
warheadsWaitGroup *sync.WaitGroup, fireCmdCh <-chan time.Time, results chan <-*Damage) {

    defer warheadsWaitGroup.Done()
    for fc := range fireCmdCh {
        results <- missile.hit(target, fc)
    }

}

// Hit the Target
func (missile *Missile) hit(target *Target, fireCmdTime time.Time) *Damage {

    damage := &Damage{timestamp: fireCmdTime}
    req, err := target.request()
    if err != nil {
        return damage
    }

    damage.startTime = time.Now()
    // Do http request
    resp, err := missile.client.Do(req)

    // Calculate the latency
    damage.endTime = time.Now()
    damage.latency = damage.endTime.Sub(damage.startTime)

    if err != nil {
        return damage
    }
    defer resp.Body.Close()

    // Just discard the response body
    in, err := io.Copy(ioutil.Discard, resp.Body)
    if err != nil {
        return damage
    }
    // Calculate the bytes received
    damage.receivedBytes = uint64(in)

    // Calculate the bytes sent
    if req.ContentLength != -1 {
        damage.sentBytes = uint64(req.ContentLength)
    }
    // Calculate the err info
    if damage.statusCode = resp.StatusCode; damage.statusCode < 200 || damage.statusCode >= 400 {
        damage.error = resp.Status
    }
    return damage
}

// Stop stops the current attack.
func (missile *Missile) stop(boomOpts *BoomOptions) {
    log.Println("Missle will stop.")
    select {
    case <-missile.stopAttack:
        return
    default:
        close(missile.stopAttack)
    }

}

// Judge which time is later
func max(a, b time.Time) time.Time {
    if a.After(b) {
        return a
    }
    return b
}
