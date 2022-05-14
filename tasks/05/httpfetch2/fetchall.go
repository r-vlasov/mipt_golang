package httpfetch2

import (
        "context"
        "bytes"
        "net"
        "sync"
        "net/http"
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


func FetchAll(ctx context.Context, c *http.Client, requests <-chan Request) <-chan Result {
        var wg sync.WaitGroup
        ch := make(chan Result)

        // main httprequest worker
        worker := func(request *Request) {
                defer wg.Done()
                var result Result

                reqAssembler, err := http.NewRequestWithContext(
                        ctx,
                        request.Method,
                        request.URL,
                        bytes.NewReader(request.Body),
                )
                if err != nil {
                        return
                }

                response, err := c.Do(reqAssembler)
                if err != nil {
                        // context deadline exceeded (Client.Timeout exceeded while awaiting headers) preventing
                        if err, ok := err.(net.Error); ok && err.Timeout() {
                                return
                        }
                        result.Error = err
                        ch <- result
                        return
                }
                result.StatusCode = response.StatusCode
                ch <- result
                response.Body.Close()
        }

        processing := true
        go func() {
                for processing {
                        select {
                        case request, ok := <- requests:
                                processing = ok
                                wg.Add(1)
                                go worker(&request)
                        case <- ctx.Done():
                                processing = false
                        }
                }
                wg.Wait()
                close(ch)
        }()
        return ch
}

