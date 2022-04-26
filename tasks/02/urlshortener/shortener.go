package urlshortener

import (
	"net/http"
	"net/url"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/go-chi/chi"
)

type URLShortener struct {
	addr string
	mapping map[string]string
}

func NewShortener(addr string) *URLShortener {
	return &URLShortener{
		addr: addr,
		mapping: map[string]string{},
	}
}

func (s *URLShortener) HandleSave(rw http.ResponseWriter, req *http.Request) {
	raw_url := req.URL.Query().Get("u")
	_, err := url.Parse(raw_url)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
	}
	hasher := md5.New()
	hasher.Write([]byte(raw_url))
	mapped_path := hex.EncodeToString(hasher.Sum(nil))
	if _, ok := s.mapping[mapped_path]; ok {
		rw.WriteHeader(http.StatusInternalServerError)
	} else {
		s.mapping[mapped_path] = raw_url
		_, err := fmt.Fprintf(rw, "%s/%s", s.addr, mapped_path)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (s *URLShortener) HandleExpand(rw http.ResponseWriter, req *http.Request) {
	r_url := chi.URLParam(req, "key")
	if mapped_path, ok := s.mapping[r_url]; ok {
		http.Redirect(rw, req, mapped_path, http.StatusMovedPermanently)
	} else {
		rw.WriteHeader(http.StatusNotFound)
	}
}
