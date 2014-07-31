package hipache

import (
	"github.com/3onyc/hipdate"
	"github.com/garyburd/redigo/redis"
	"log"
)

type HipacheBackend struct {
	r redis.Conn
}

func NewHipacheBackend(r redis.Conn) *HipacheBackend {
	return &HipacheBackend{
		r: r,
	}
}

func (hb *HipacheBackend) AddUpstream(
	h hipdate.Host,
	u hipdate.Upstream,
) error {
	exists, err := hb.hostExists(h)
	if err != nil {
		log.Println(err)
	}

	if !exists {
		if err := hb.hostCreate(h); err != nil {
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
func (hb *HipacheBackend) hostExists(h hipdate.Host) (bool, error) {
	return redis.Bool(hb.r.Do("EXISTS", h.Key()))
}

func (hb *HipacheBackend) hostDelete(h hipdate.Host) error {
	if _, err := hb.r.Do("DEL", h.Key()); err != nil {
		return err
	}
	log.Println("Deleted", h)

	return nil
}

func (hb *HipacheBackend) hostCreate(h hipdate.Host) error {
	if _, err := hb.r.Do("RPUSH", h.Key(), h); err != nil {
		return err
	}
	log.Println("Created", h)

	return nil
}

func (hb *HipacheBackend) hostClear(h hipdate.Host) error {
	if err := hb.hostDelete(h); err != nil {
		return err
	}

	if err := hb.hostCreate(h); err != nil {
		return err
	}

	log.Println("Initialised", h)
	return nil
}
