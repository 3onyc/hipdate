package hipache

import (
	"errors"
	"github.com/3onyc/hipdate/backends"
	"github.com/3onyc/hipdate/shared"
	"github.com/garyburd/redigo/redis"
	"log"
)

var (
	MissingRedisUrlError = errors.New("REDIS_URL not specified")
)

type HipacheBackend struct {
	r redis.Conn
}

func NewHipacheBackend(opts shared.OptionMap) (backends.Backend, error) {
	ru, ok := opts["REDIS_URL"]
	if !ok {
		return nil, MissingRedisUrlError
	}

	r, err := createRedisConn(ru)
	if err != nil {
		return nil, err
	}

	return &HipacheBackend{
		r: *r,
	}, nil
}

func (hb *HipacheBackend) AddUpstream(
	h shared.Host,
	u shared.Upstream,
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
	h shared.Host,
	u shared.Upstream,
) error {
	if _, err := hb.r.Do("LREM", h.Key(), 0, u); err != nil {
		return err
	}

	log.Println("Unregistered", h, u)
	return nil
}
func (hb *HipacheBackend) hostExists(h shared.Host) (bool, error) {
	return redis.Bool(hb.r.Do("EXISTS", h.Key()))
}

func (hb *HipacheBackend) hostDelete(h shared.Host) error {
	if _, err := hb.r.Do("DEL", h.Key()); err != nil {
		return err
	}
	log.Println("Deleted", h)

	return nil
}

func (hb *HipacheBackend) hostCreate(h shared.Host) error {
	if _, err := hb.r.Do("RPUSH", h.Key(), h); err != nil {
		return err
	}
	log.Println("Created", h)

	return nil
}

func (hb *HipacheBackend) hostClear(h shared.Host) error {
	if err := hb.hostDelete(h); err != nil {
		return err
	}

	if err := hb.hostCreate(h); err != nil {
		return err
	}

	log.Println("Initialised", h)
	return nil
}

func init() {
	backends.BackendMap["hipache"] = NewHipacheBackend
}
