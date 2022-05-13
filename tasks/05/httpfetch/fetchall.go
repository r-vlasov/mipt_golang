package httpfetch

import (
	"bytes"
	"sync"
	"net/http"
)

type Request struct {
	Method	string
	URL	string
	Body	[]byte
}

type Result struct {
	StatusCode	int
	Error		error
}

func FetchAll(c *http.Client, requests []Request) []Result {

	var wg sync.WaitGroup
	quantityReq := len(requests)
	wg.Add(quantityReq)
	results := make([]Result, quantityReq)

	// main worker to goroutine that create request and fill results array
	requestWorker := func(i int) {
		defer wg.Done()
		reqHttp, err := http.NewRequest(
			requests[i].Method, requests[i].URL, bytes.NewReader(requests[i].Body),
		)
		if err != nil {
			results[i].Error = err
			return
		}

		response, err := c.Do(reqHttp)
		if err != nil {
			results[i].Error = err
			return
		}
		results[i].StatusCode = response.StatusCode
		defer response.Body.Close()
	}

	for i := 0; i < quantityReq; i++ {
		go requestWorker(i)
	}
	wg.Wait()
	return results
}
