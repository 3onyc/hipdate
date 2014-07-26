package backends

import (
	"github.com/3onyc/hipdate"
	"github.com/garyburd/redigo/redis"
	"log"
)

type HipacheBackend struct {
	r redis.Conn
}

func (hb *HipacheBackend) AddUpstream(
	h hipdate.Host,
	u hipdate.Upstream,
) error {
	exists, err := hb.HostExists(h)
	if err != nil {
		log.Println(err)
	}

	if !exists {
		if err := hb.HostCreate(h); err != nil {
			log.Println(err)
		}
	}

	if _, err := hb.r.Do("RPUSH", h.Key(), u); err != nil {
		return err
	}
	log.Println("Registered", h, u)

	return nil
}
func (hb *HipacheBackend) RemoveUpstream(
	h hipdate.Host,
	u hipdate.Upstream,
) error {
	if _, err := hb.r.Do("LREM", h.Key(), 0, u); err != nil {
		return err
	}

	log.Println("Unregistered", h, u)
	return nil
}
func (hb *HipacheBackend) HostExists(h hipdate.Host) (bool, error) {
	return redis.Bool(hb.r.Do("EXISTS", h.Key()))
}

func (hb *HipacheBackend) HostDelete(h hipdate.Host) error {
	if _, err := hb.r.Do("DEL", h.Key()); err != nil {
		return err
	}
	log.Println("Deleted", h)

	return nil
}

func (hb *HipacheBackend) HostCreate(h hipdate.Host) error {
	if _, err := hb.r.Do("RPUSH", h.Key(), h); err != nil {
		return err
	}
	log.Println("Created", h)

	return nil
}

func (hb *HipacheBackend) HostInitialise(h hipdate.Host) error {
	if err := hb.HostDelete(h); err != nil {
		return err
	}

	if err := hb.HostCreate(h); err != nil {
		return err
	}

	log.Println("Initialised", h)
	return nil
}

func (hb *HipacheBackend) Initialise(hl hipdate.HostList) {
	for h, ul := range hl {
		err := hb.HostInitialise(h)
		if err != nil {
			log.Println(err)
			continue
		}

		for _, u := range ul {
			err := hb.AddUpstream(h, u)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func NewHipacheBackend(r redis.Conn) *HipacheBackend {
	return &HipacheBackend{
		r: r,
	}
}
