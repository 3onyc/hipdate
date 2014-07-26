package main

import (
	"github.com/crosbymichael/skydock/docker"
	"github.com/garyburd/redigo/redis"
	"log"
	"sync"
)

type IPAddress string
type ContainerID string
type IPMap map[ContainerID]IPAddress
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

type Upstream string

func (u Upstream) Register(r redis.Conn, h Host) error {
	_, err := r.Do("RPUSH", h.Key(), u)
	if err != nil {
		return err
	}
	log.Println("Registered", h, u)

	return nil
}
func (u Upstream) Unregister(r redis.Conn, h Host) error {
	_, err := r.Do("LREM", h.Key(), 0, u)
	if err != nil {
		return err
	}
	log.Println("Unregistered", h, u)

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

type UpstreamList []Upstream
type HostList map[Host]UpstreamList

func (hl HostList) Add(h Host, u Upstream) {
	hl[h] = append(hl[h], u)
}
func (hl HostList) Initialise(r redis.Conn) {
	for h, ul := range hl {
		err := h.Initialise(r)
		if err != nil {
			log.Println(err)
			continue
		}

		for _, u := range ul {
			err := u.Register(r, h)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func getHostnames(c *docker.Container) []Host {
	env := parseEnv(c.Config.Env)
	hosts := []Host{}

	if _, exists := env["WEB_HOSTNAME"]; exists {
		for _, host := range parseHostnameVar(env["WEB_HOSTNAME"]) {
			hosts = append(hosts, Host(host))
		}
	}

	return hosts
}
