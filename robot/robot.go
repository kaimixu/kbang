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
	"net"
	"io"
	"fmt"
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

type request struct {
	preq 	*http.Request
	conf 	*RequestConf
}

type Roboter struct {
	n 		int
	c 		int
	weight 		int
	requests 	[]request
	httpConf 	*HttpConf

	output 		*output
}

func NewRoboter(n, c int, httpConf *HttpConf) *Roboter{
	return &Roboter{
		n:		n,
		c:		c,
		httpConf:	httpConf,
		requests:	make([]request, 0),
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
			this.requests = append(this.requests, request{preq:r, conf:&reqC})
		}

		this.weight += reqC.Weight
	}

	return  nil
}

func (this *Roboter) Run() {
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)

	fmt.Fprintf(os.Stdout, "start...\n")
	st := time.Now()
	go func() {
		<-s
		fmt.Println("receive  sigint\n")
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
		go func() {
			this.startWorker(this.n / this.c)
			wg.Done()
		}()
	}
	wg.Wait()
}

func (this *Roboter) startWorker(n int) {
	tr := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   time.Duration(this.httpConf.Timeout) * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		DisableKeepAlives: !this.httpConf.KeepAlive,
	}
	client := &http.Client{Transport: tr}

	for i := 0; i < n; i++ {
		req := &this.requests[i%this.weight]
		this.sendRequest(client, req)
	}
}

func (this *Roboter) sendRequest(c *http.Client, req *request) {
	s := time.Now()
	var code int

	resp, err := c.Do(cloneRequest(req.preq, req.conf.PostData))
	if err == nil {
		code = resp.StatusCode
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
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
