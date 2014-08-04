package hipdate

import (
	"fmt"
	"github.com/3onyc/hipdate/backends"
	"github.com/3onyc/stoppableListener"
	"log"
	"net"
	"net/http"
)

type HttpServer struct {
	l *stoppableListener.StoppableListener
	s *http.Server
	b backends.Backend
}

func NewHttpServer(b backends.Backend) *HttpServer {
	return &HttpServer{
		s: &http.Server{},
		b: b,
	}
}

func (h *HttpServer) Start() error {
	l, err := net.Listen("tcp", ":8889")
	if err != nil {
		return err
	}

	sl, err := stoppableListener.New(l)
	if err != nil {
		return err
	}

	http.HandleFunc("/status", h.status)

	h.l = sl
	h.s.Serve(h.l)

	return nil
}

func (h *HttpServer) Stop() {
	h.l.Stop()
	log.Println("[http] stopped")
}

func (h *HttpServer) status(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Add("Content-Type", "text/html")
	fmt.Fprintf(rw, h.b.ListHosts().Pprint())
}
