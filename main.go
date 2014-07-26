package main

import (
	"github.com/crosbymichael/skydock/docker"
	"github.com/garyburd/redigo/redis"
	"log"
	"sync"
)

type IPMap map[string]string
type Application struct {
	Redis  redis.Conn
	Docker docker.Docker
	Hosts  HostList
	Status sync.WaitGroup
	IPs    IPMap
}

func NewApplication(
	r redis.Conn,
	d docker.Docker,
) *Application {
	return &Application{
		Redis:  r,
		Docker: d,
		IPs:    IPMap{},
	}
}

type Backend string

func (b Backend) Register(r redis.Conn, h Host) error {
	_, err := r.Do("RPUSH", h.Key(), b)
	if err != nil {
		return err
	}
	log.Println("Registered", h, b)

	return nil
}
func (b Backend) Unregister(r redis.Conn, h Host) error {
	_, err := r.Do("LREM", h.Key(), 0, b)
	if err != nil {
		return err
	}
	log.Println("Unregistered", h, b)

	return nil
}

type Host string

func (h Host) Exists(r redis.Conn) (bool, error) {
	exists, err := redis.Bool(r.Do("EXISTS", h.Key()))
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (h Host) Delete(r redis.Conn) error {
	if _, err := r.Do("DEL", h.Key()); err != nil {
		return err
	}
	log.Println("Deleted", h)

	return nil
}

func (h Host) Create(r redis.Conn) error {
	if _, err := r.Do("RPUSH", h.Key(), h); err != nil {
		return err
	}
	log.Println("Created", h)

	return nil
}

func (h Host) Initialise(r redis.Conn) error {
	if err := h.Delete(r); err != nil {
		return err
	}

	if err := h.Create(r); err != nil {
		return err
	}
	log.Println("Initialised", h)

	return nil
}

func (h Host) Key() string {
	return "frontend:" + string(h)
}

type BackendList []Backend
type HostList map[Host]BackendList

func (hl HostList) Add(h Host, b Backend) {
	hl[h] = append(hl[h], b)
}
func (hl HostList) Initialise(r redis.Conn) {
	for h, bl := range hl {
		err := h.Initialise(r)
		if err != nil {
			log.Println(err)
			continue
		}

		for _, b := range bl {
			err := b.Register(r, h)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
