package middleware

import (
	"net/http"
	"sync"
)

func Limit(l Limiter) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if !l.TryAcquire() {
				w.WriteHeader(http.StatusTooManyRequests)
			} else {
				defer l.Release()
			}
			h.ServeHTTP(w, req)
		})
	}
}

type Limiter interface {
	TryAcquire() bool
	Release()
}

type MutexLimiter struct {
	mutex sync.Mutex
	cnt int
	cnt_limit int
}

func NewMutexLimiter(count int) *MutexLimiter {
	ml := new(MutexLimiter)
	ml.cnt_limit = count
	ml.cnt = 0
	return ml
}

func (l *MutexLimiter) TryAcquire() bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if l.cnt < l.cnt_limit {
		l.cnt += 1
		return true
	}
	return false
}

func (l *MutexLimiter) Release() {
	defer l.mutex.Unlock()
	l.mutex.Lock()
	l.cnt -= 1
}

type ChanLimiter struct {
	ch      chan bool
	cnt int
	cnt_limit   int
}

func NewChanLimiter(count int) *ChanLimiter {
	cl := new(ChanLimiter)
	cl.ch = make(chan bool, 1)
	cl.ch <- true
	cl.cnt = 0
	cl.cnt_limit = count
	return cl
}

func (l *ChanLimiter) TryAcquire() bool {
	<-l.ch
	defer func() {
		l.ch <- true
	}()
	if l.cnt < l.cnt_limit {
		l.cnt += 1
		return true
	}
	return false
}

func (l *ChanLimiter) Release() {
	<-l.ch
	defer func() {
		l.ch <- true
	}()
	l.cnt -= 1
}
