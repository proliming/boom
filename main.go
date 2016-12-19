package main

import (
    "flag"
    "runtime"
    "os"
    "fmt"
    "errors"
    "time"
)

const BoomVersion = "1.0"

var (
    // -cpu:要使用的cpu个数, 默认使用1个CPU
    cpuToUse int
    // -h: 显示使用帮助
    showHelpUsage bool
    // -l: 开启log输出， 默认不显示log信息
    showLogs bool
    // -V: boom 版本
    showVersion bool
)
var (
    errZeroRate = errors.New("每秒请求次数必须大于0")
    errBadCert = errors.New("认证失败")
    errHeaders = errors.New("请求头格式不正确")
)

type BoomOptions struct {
    // -A: Authentication,基础认证  auth-username:password
    authentication             string
    // -C: 在请求时发送Cookie   cookie-name=value
    requestCookies             string
    // -c: Content-Type, post数据时的Content—Type
    requestPostDataContentType string
    // -D: post 的数据，支持文件和 引号中的内容
    requestPostData            string
    // -g: 开启的goroutines个数(线程数)，默认1个
    requestGoroutines          int
    // -H: 自定义附加的请求头  head-type:value
    requestHeaders             string
    // -k: 启用KeepAlive功能，允许在一个http会话中发送多个requests 默认关闭
    enableKeepAlive            bool

    // -la: local address, 建立连接时，使用的本地地址
    localAddress               string
    // -m: 自定义请求方法，默认GET
    requestMethod              string
    // -t: 压测持续时间
    requestDuration            time.Duration
    // -u： 请求的url
    requestUrl                 string
    // -o: 结果输出到指定位置
    resultOutput               string
    // -R: 生成结果报告text, json, plot
    generateReports            string
    // -r: 每秒请求次数
    requestPerSec              int
    // -P: Protocol,请求使用的协议
    requestProtocol            string
    // -s: timeout, 一次请求超时时间设置 单位为s
    requestTimeout             time.Duration
}


// 进一步处理命令行参数，例如请求头，认证等参数
func parseArgs() (*BoomOptions, error) {
    boomOpts := &BoomOptions{
        requestGoroutines:1000,
        requestDuration:1,
        enableKeepAlive:false,
        requestTimeout:30,
    }
    flag.StringVar(&boomOpts.authentication, "A", "", "Authentication,基础认证  auth-username:password")
    flag.StringVar(&boomOpts.requestCookies, "C", "", "在请求时发送Cookie, cookie-name=value;cookie-name2=value2")
    flag.IntVar(&cpuToUse, "cpu", 1, "使用的CPU个数")
    flag.StringVar(&boomOpts.requestPostDataContentType, "c", "", "Content-Type, post数据时的Content—Type")
    flag.StringVar(&boomOpts.requestPostData, "D", "", "Post 的数据，支持文件和 引号中的内容, 如果是文件要以@开头，" +
        "例如 '@/home/work/aa.json'")
    flag.IntVar(&boomOpts.requestGoroutines, "g", 1000, "开启的goroutines个数(线程数)，默认1个")
    flag.StringVar(&boomOpts.requestHeaders, "H", "", "自定义附加的请求头  head-type:value")
    flag.BoolVar(&boomOpts.enableKeepAlive, "k", false, "启用KeepAlive功能，允许在一个http会话中发送多个requests 默认关闭")
    flag.BoolVar(&showLogs, "l", false, "开启log输出,默认不显示log信息")
    flag.StringVar(&boomOpts.localAddress, "la", "", "建立连接时，使用的本地地址")
    flag.StringVar(&boomOpts.requestMethod, "m", "GET", "自定义请求方法，默认GET")
    flag.DurationVar(&boomOpts.requestDuration, "t", time.Second, "压测持续时间")
    flag.StringVar(&boomOpts.requestUrl, "u", "", "请求的url")
    flag.StringVar(&boomOpts.resultOutput, "o", "", "结果输出到指定位置")
    flag.StringVar(&boomOpts.requestProtocol, "P", "HTTP", "Protocol,请求使用的协议")
    flag.DurationVar(&boomOpts.requestTimeout, "s", 30 * time.Second, "一次请求超时时间设置,单位为s")
    flag.IntVar(&boomOpts.requestPerSec, "r", 50, "每秒请求的次数")
    flag.StringVar(&boomOpts.generateReports, "R", "", "生成结果报告[text, json, plot]")
    flag.BoolVar(&showVersion, "V", false, "显示当前Boom版本")
    flag.Parse()
    //TODO 进一步处理
    return boomOpts, nil
}


// 显示帮助信息
func usage() {
    usage := `aaaa`
    fmt.Print(usage)
    os.Exit(0)
}

// Main entrance
func main() {
    boomOpts, err := parseArgs()
    if err != nil {
        panic("参数错误:" + err.Error())
        os.Exit(1)
    }
    if showVersion {
        fmt.Println(BoomVersion)
        os.Exit(0)
    }
    if showHelpUsage {
        usage()
        os.Exit(0)
    }
    // set GOMAXPROCS
    runtime.GOMAXPROCS(cpuToUse)
    boom(boomOpts)
}
