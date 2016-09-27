package robot

import (
	"time"
	"os"
	"os/signal"
	"net/http"
	"strings"
	"errors"
	"io/ioutil"
	"sync"
	"fmt"
	"net"
)

type RequestConf struct {
	Weight		int	`weight`
	Method 		string	`method`
	Url 		string	`url`
	ContentType	string	`content_type`
	PostData	string	`post_data`
}

type HttpConf struct {
	KeepAlive 	bool		`keepalive`
	Header 		string		`header`
	Timeout		int		`timeout`
	Request 	[10]RequestConf `request`
}

type Roboter struct {
	n 		int
	c 		int
	weight 		int
	httpConf 	*HttpConf

	hc 		*http.Client
	requests	[]*http.Request
	output 		*output
}

func NewRoboter(n, c int, httpConf *HttpConf) *Roboter{
	return &Roboter{
		n:		n,
		c:		c,
		httpConf:	httpConf,
		requests:	make([]*http.Request, 0),
		output:		newOutput(n, c),
	}
}

func (this *Roboter) CreateRequest() error {
	var header []string
	if this.httpConf.Header != "" {
		header = strings.SplitN(this.httpConf.Header, ":", 2)
		if len(header) != 2 {
			return errors.New("invalid http header")
		}
	}

	type request struct {
		preq 	*http.Request
		conf 	*RequestConf
	}
	reqs := make([]*request, 0)
	for _, reqC := range this.httpConf.Request {
		method := strings.ToUpper(reqC.Method)
		if method == "" {
			continue
		}
		if method != "GET" && method != "POST" {
			return errors.New("invalid http method")
		}

		r, err := http.NewRequest(method, reqC.Url, nil)
		if err != nil {
			return err
		}
		if len(reqC.ContentType) != 0 {
			r.Header.Set("Content-Type", reqC.ContentType)
		}
		if len(header) == 2 {
			r.Header.Add(header[0], header[1])
		}

		for i := 0; i < reqC.Weight; i++ {
			reqs = append(reqs, &request{preq:r, conf:&reqC})
		}

		this.weight += reqC.Weight
	}

	for i := 0; i < this.n; i++ {
		req := reqs[i%this.weight]
		this.requests = append(this.requests, cloneRequest(req.preq, req.conf.PostData))
	}

	return  nil
}

func (this *Roboter) Run() {
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)

	tr := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   time.Duration(this.httpConf.Timeout) * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		DisableKeepAlives: !this.httpConf.KeepAlive,
	}
	this.hc = &http.Client{Transport: tr}

	fmt.Println("start...")
	st := time.Now()
	go func() {
		<-s
		fmt.Println("receive  sigint")
		this.output.finalize(time.Now().Sub(st).Seconds())
		os.Exit(1)
	}()

	this.startWorkers()
	this.output.finalize(time.Now().Sub(st).Seconds())
}

func (this *Roboter) startWorkers() {
	var wg sync.WaitGroup
	wg.Add(this.c)

	for i := 0; i < this.c; i++ {
		go func(rid int) {
			this.startWorker(rid, this.n / this.c)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func (this *Roboter) startWorker(rid, num int) {
	for i := 0; i < num; i++ {
		req := this.requests[rid*num+i]
		this.sendRequest(req)
	}
}

func (this *Roboter) sendRequest(req *http.Request) {
	s := time.Now()
	var code int

	resp, err := this.hc.Do(req)
	if err == nil {
		code = resp.StatusCode
	}

	this.output.addResult(&result{
		statusCode:    code,
		duration:      time.Now().Sub(s),
	})
}

func cloneRequest(r *http.Request, body string) *http.Request {
	r2 := new(http.Request)
	*r2 = *r
	r2.Header = make(http.Header, len(r.Header))
	for k, s := range r.Header {
		r2.Header[k] = append([]string(nil), s...)
	}
	r2.Body = ioutil.NopCloser(strings.NewReader(body))

	return r2
}
