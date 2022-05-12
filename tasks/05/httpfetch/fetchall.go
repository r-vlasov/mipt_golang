package httpfetch

import (
	"net/http"
	"sync"
	"bytes"
	"log"
)

type Request struct {
	Method string
	URL    string
	Body   []byte
}

type Result struct {
	StatusCode int
	Error      error
}

func FetchAll(c *http.Client, requests []Request) []Result {
	var wg sync.WaitGroup
	quantityReq := len(requests)
	results := make([]Result, quantityReq)

	handleRequest := func(i int) {
		req, err := http.NewRequest(
			requests[i].Method, 
			requests[i].URL, 
			bytes.NewReader(requests[i].Body),
		)
		if err != nil {
			log.Printf("can't create a request struct: %v", err)
			return
		}
		response, err := c.Do(req)
		if err != nil {
			results[i].Error = err
		}else {
			results[i] = Result {
				StatusCode: 	response.StatusCode,
				Error:		err,
			}
		}
		return

		defer response.Body.Close()
	}
	for i := 0; i < quantityReq; i++ {
		go handleRequest(i)
	}
	wg.Wait()
	return results

}
