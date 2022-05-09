package middleware

import (
	"log"
	"net/http"
)

func Recover(logger *log.Logger) func(http.Handler) http.Handler {
   return func(h http.Handler) http.Handler {
       return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
           defer func() {
               r := recover()
               if r != nil {
                   logger.Printf("[ERROR] Panic caught: %v", r)
		   w.WriteHeader(http.StatusInternalServerError)
		   w.Write([]byte("Internal Server Error\n"))
               }
           }()
           h.ServeHTTP(w, req)
       })
   }
}
