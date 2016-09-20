package main

import (
	"fmt"
	"runtime"
	"flag"
	"os"
	"github.com/kaimixu/kbang/conf"
	"github.com/kaimixu/kbang/robot"
)

var (
	n,c,t		int
	keepalive	bool
	url 		string
	requestBody	string
	contentType	string
	header		string

	cfgFile 	string
)

var usage =
`Usage: kbang [options...] <url>             (1st form)
    or: kbang [options...] -f configfile    (2st form)

options:
    -n  Number of requests to run (default: 10)
    -c  Number of requests to run concurrency (default: 1)
    -t  Request connection timeout in second (default: 1s)
    -H  Http header, eg. -H "Host: www.example.com"
    -k[=true|false]  Http keep-alive (default: true)
    -d  Http request body to POST
    -T  Content-type header to POST, eg. 'application/x-www-form-urlencoded'
        (Default：text/plain)

`

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}
	flag.IntVar(&n, "n", 10, "")
	flag.IntVar(&c, "c", 1, "")
	flag.IntVar(&t, "t", 1, "")
	flag.BoolVar(&keepalive, "k", true, "")
	flag.StringVar(&cfgFile, "f", "", "")
	flag.StringVar(&requestBody, "d", "", "")
	flag.StringVar(&contentType, "T", "text/plain", "")
	flag.StringVar(&header, "H", "", "")
	flag.Parse()

	if flag.NArg() < 1 && len(cfgFile) == 0 {
		abort("")
	}

	method := "GET"
	if requestBody != "" {
		method = "POST"
	}

	var httpConf = robot.HttpConf{
		KeepAlive: keepalive,
		Header: header,
		Timeout: t,
	}
	if flag.NArg() > 0 {
		url = flag.Args()[0]
		httpConf.Request[0] = robot.RequestConf{
			Weight: 	1,
			Method: 	method,
			Url: 		url,
			ContentType: 	contentType,
			PostData:	requestBody,
		}
	}else {
		cfg := conf.NewConf()
		err := cfg.LoadFile(cfgFile)
		if err != nil {
			abort(err.Error())
		}
		err = cfg.Parse(&httpConf)
		if err != nil {
			abort(err.Error())
		}
	}

	robot := robot.NewRoboter(n, c, &httpConf)
	err := robot.CreateRequest()
	if err != nil {
		abort(err.Error())
	}

	robot.Run()
}

func abort(errmsg string) {
	if errmsg != "" {
		fmt.Fprintf(os.Stderr, "%s\n\n", errmsg)
	}

	flag.Usage()
	os.Exit(1)
}