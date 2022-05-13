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
	for i := 0; i < quantityReq; i++ {
		go requestWorker(c, &wg, &requests[i], &results[i])
	}
	wg.Wait()
	return results
}

func requestWorker(c *http.Client, wg *sync.WaitGroup, request *Request, result *Result) {
	reqHttp, err := http.NewRequest(request.Method, request.URL, bytes.NewReader(request.Body))
	if err != nil {
		return
	}

	response, err := c.Do(reqHttp)
	if err != nil {
		result.Error = err
		return
	}

	defer wg.Done()
	defer response.Body.Close()
	result.StatusCode = response.StatusCode
}
