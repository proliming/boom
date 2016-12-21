# boom [![Build Status](https://travis-ci.org/proliming/boom.svg?branch=master)](https://travis-ci.org/proliming/boom)

Boom is a HTTP load/stress testing tool.

![](boom-logo.png)

## Usage manual
```console
Usage of ./boom:
  -A string
        Supply BASIC Authentication credentials to the server. The username and password are separated by a single : .
  -C string
        Add a Cookie: line to the request like: cookie-name=value
  -D string
        File or just a string containing data to POST. Remember to also set -c.When using a file for input, remember add '@@' prefix to the file path. eg. @@/home/work/a.json
  -H string
        Append extra headers to the request like: head-type:value
  -P string
        Specify SSL/TLS protocol . (default "HTTP")
  -R string
        Generate reports in [text, json, plot]
  -V     Show version of boom then exit
  -c string
        Content-type header to use for POST/PUT data, eg. application/x-www-form-urlencoded. Default is text/plain.
  -cpu int
        The cpu to use when sending requests (default 1)
  -g int
         Number of threads(goroutines) to perform for the test. (default 100)
  -k    Enable the HTTP KeepAlive feature
  -l    Enable log output
  -la string
        Local address  to bind to when making outgoing connections.
  -m string
        Custom HTTP method for the requests. (default "GET")
  -n int
        Number of requests to perform for the test. If this flag > 0, the -t and -r will be ignore.
  -o string
        Output the reports in specified location (default "Stdout")
  -r int
        Number of requests to perform at one sec. (default 50)
  -s duration
        Maximum number of seconds to wait before a request times out. (default 30s)
  -t duration
        Duration of this test. (default 1s)
  -u string
        The url to request
```