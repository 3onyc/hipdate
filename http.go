package hipdate

import (
	"encoding/json"
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

	http.HandleFunc("/api/v1/status.json", h.status)

	h.l = sl
	h.s.Serve(h.l)

	return nil
}

func (h *HttpServer) Stop() {
	h.l.Stop()
	log.Println("NOTICE [http] stopped")
}

func (h *HttpServer) status(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Add("Content-Type", "application/json")
	hs, err := h.b.ListHosts()
	if err != nil {
		rw.WriteHeader(500)
		fmt.Fprint(rw, err)
		return
	}

	b, err := json.MarshalIndent(hs, "", "    ")
	if err != nil {
		rw.WriteHeader(500)
		fmt.Fprint(rw, err)
		return
	}

	if _, err := rw.Write(b); err != nil {
		log.Println("ERROR [http]", err)
	}
}
