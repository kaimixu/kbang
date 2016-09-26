package robot

import (
	"fmt"
	"time"
	"os"
)

type result struct {
	statusCode    int
	duration      time.Duration
}

type output struct {
	average  float64
	rps      float64

	n 			int
	c 			int
	reqNumTotal 		int64
	reqNumFail		int64
	reqNumSucc		int64
	costTimeTotal 		float64
	results chan *result
}

func newOutput(n int, c int) *output {
	return &output{
		n: n,
		c: c,
		results: make(chan *result, n),
	}
}

func (this *output) addResult(res *result) {
	this.results <- res
}

func (this *output) finalize(costTime float64) {
	this.costTimeTotal = costTime

	for {
		select {
		case res := <-this.results:
			this.reqNumTotal++
			if res.statusCode != 200 {
				this.reqNumFail++
			}else {
				this.reqNumSucc++
			}
		default:
			this.rps = float64(this.reqNumTotal) / this.costTimeTotal
			this.average = this.costTimeTotal / float64(this.reqNumTotal)
			this.print()
			return
		}
	}
}

func (this *output) print() {
	if this.reqNumTotal > 0 {
		fmt.Printf("Summary:\n")
		fmt.Printf("  Concurrency Level:\t%d\n", this.c)
		fmt.Printf("  Time taken for tests:\t%4.4f secs\n", this.costTimeTotal)
		fmt.Printf("  Complete requests:\t%d\n", this.reqNumTotal)
		fmt.Printf("  Failed requests:\t%d\n", this.reqNumFail)
		fmt.Printf("  Success requests:\t%d\n", this.reqNumSucc)
		fmt.Printf("  Requests per second:\t%4.4f\n", this.rps)
		fmt.Printf("  Average time per request:\t%4.4f\n", this.average)
	}

	os.Exit(0)
}