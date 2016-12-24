package main

import (
    "net"
    "net/http"
    "time"
    "crypto/tls"
    "sync"
    "io"
    "io/ioutil"
    "log"
)

const (
    defaultTimeout = 30 * time.Second
    defaultMaxIdleConnections = 100
    defaultWarheads = 100
    noFollow = -1
)
// Missile is a wrapper of http.Client and some properties
// A missile can carry many warheads means multi goroutines
type Missile struct {
    ctrl   *CtrlCenter
    dialer *net.Dialer
    client http.Client
}

type CtrlCenter struct {
    Timeout            time.Duration
    Warheads           int // How many warhead can this missile carry
    MaxIdleConnections int
    KeepAlive          time.Duration
    Http2Enable        bool
    MaxRedirects       int
    LocalAddr          *net.IPAddr
    TLSConfig          *tls.Config
    Cancel             chan struct{}
}

var (
    defaultLocalAddr = &net.IPAddr{IP: net.IPv4zero}
    defaultTLSConfig = &tls.Config{InsecureSkipVerify: true}
)

// Create a default control center.
func NewDefaultCtrlCenter() *CtrlCenter {
    c := &CtrlCenter{}
    c.Timeout = defaultTimeout
    c.MaxIdleConnections = defaultMaxIdleConnections
    c.Warheads = defaultWarheads
    c.KeepAlive = 0
    c.Http2Enable = false
    c.LocalAddr = defaultLocalAddr
    c.TLSConfig = defaultTLSConfig
    c.Cancel = make(chan struct{})
    return c;
}

// Create a missile with default options.
func NewMissile() *Missile {
    return NewCustomMissile(nil)
}

// Create a missile with custom options
func NewCustomMissile(ct *CtrlCenter) *Missile {
    missile := &Missile{}
    if ct == nil {
        ct = NewDefaultCtrlCenter()
    }
    missile.ctrl = ct
    missile.dialer = &net.Dialer{
        LocalAddr:  &net.TCPAddr{IP: defaultLocalAddr.IP, Zone: defaultLocalAddr.Zone},
        KeepAlive: ct.KeepAlive,
        Timeout:   ct.Timeout,
    }

    missile.client = http.Client{
        Transport: &http.Transport{
            Proxy: http.ProxyFromEnvironment,
            Dial:  missile.dialer.Dial,
            ResponseHeaderTimeout: defaultTimeout,
            TLSClientConfig:       defaultTLSConfig,
            TLSHandshakeTimeout:   10 * time.Second,
            MaxIdleConnsPerHost:   ct.MaxIdleConnections,
        },
    }
    return missile
}

// Launch the Missile
func (missile *Missile) Launch(target *Target, totalHits int, hitPerSecond int, du time.Duration) <-chan *Damage {

    var warheadsWaitGroup sync.WaitGroup
    damagesCh := make(chan *Damage)
    fireCmdCh := make(chan time.Time)
    log.Println("Fireing...")
    // Each warhead standard for a single goroutine
    for i := 0; i < missile.ctrl.Warheads; i++ {
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
                //time.Sleep(1 * time.Nanosecond)
                select {
                case fireCmdCh <- time.Now():
                    if done++; done == totalHits {
                        return
                    }
                case <-missile.ctrl.Cancel:
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
                case <-missile.ctrl.Cancel:
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

func (missile *Missile) fire(target *Target, warheadsWaitGroup *sync.WaitGroup, fireCmdCh <-chan time.Time, results chan <-*Damage) {

    defer warheadsWaitGroup.Done()
    for fc := range fireCmdCh {
        results <- missile.hit(target, fc)
    }

}

// Hit the Target
func (missile *Missile) hit(target *Target, fireCmdTime time.Time) *Damage {

    damage := &Damage{Timestamp: fireCmdTime}
    req, err := target.Request()
    if err != nil {
        return damage
    }

    damage.StartTime = time.Now()
    // Do http request
    resp, err := missile.client.Do(req)

    // Calculate the latency
    damage.EndTime = time.Now()
    damage.Latency = damage.EndTime.Sub(damage.StartTime)

    if err != nil {
        damage.Error = err.Error()
        return damage
    }
    defer resp.Body.Close()

    // Just discard the response body
    in, err := io.Copy(ioutil.Discard, resp.Body)
    if err != nil {
        damage.Error = err.Error()
        return damage
    }
    // Calculate the bytes received
    damage.ReceivedBytes = uint64(in)

    // Calculate the bytes sent
    if req.ContentLength != -1 {
        damage.SentBytes = uint64(req.ContentLength)
    }
    // Calculate the err info
    if damage.StatusCode = resp.StatusCode; damage.StatusCode != 200 {
        damage.Error = resp.Status
    }
    return damage
}

// Stop stops the current attack.
func (missile *Missile) Stop() {
    log.Println("Missle will stop.")
    select {
    case <-missile.ctrl.Cancel:
        return
    default:
        close(missile.ctrl.Cancel)
    }

}

// Judge which time is later
func max(a, b time.Time) time.Time {
    if a.After(b) {
        return a
    }
    return b
}