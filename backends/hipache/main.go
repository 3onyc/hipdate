package hipache

import (
	"errors"
	"github.com/3onyc/hipdate/backends"
	"github.com/3onyc/hipdate/shared"
	"github.com/garyburd/redigo/redis"
	"log"
)

var (
	MissingRedisUrlError = errors.New("redis url not specified")
)

type HipacheBackend struct {
	r redis.Conn
}

func NewHipacheBackend(opts shared.OptionMap) (backends.Backend, error) {
	ru, ok := opts["redis"]
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

func (hb *HipacheBackend) AddEndpoint(
	h shared.Host,
	e shared.Endpoint,
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

	if _, err := hb.r.Do("RPUSH", prefixKey(h), e.String()); err != nil {
		return err
	}
	log.Println("DEBUG [backend:hipache] Endpoint added", h, e.String())

	return nil
}
func (hb *HipacheBackend) RemoveEndpoint(
	h shared.Host,
	e shared.Endpoint,
) error {
	if _, err := hb.r.Do("LREM", prefixKey(h), 0, e.String()); err != nil {
		return err
	}

	log.Println("DEBUG [backend:hipache] Endpoint removed", h, e.String())
	return nil
}

func (hb *HipacheBackend) Initialise() error {
	return hb.clearHosts()
}

func (hb *HipacheBackend) ListHosts() (*shared.HostList, error) {
	hl := shared.HostList{}

	fe, err := hb.getFrontends()
	if err != nil {
		return nil, err
	}

	for _, f := range fe {
		r, err := redis.Values(hb.r.Do("LRANGE", f, "0", "-1"))
		if err != nil {
			return nil, err
		}

		var vs []string
		if err := redis.ScanSlice(r, &vs); err != nil {
			log.Println("ERROR [backend:hipache] ", err)
			continue
		}

		if len(vs) == 0 {
			continue
		}

		h := shared.Host(vs[0])
		hl[h] = []shared.Endpoint{}

		if len(vs) < 2 {
			continue
		}

		for _, b := range vs[1:] {
			e, err := shared.NewEndpointFromUrl(b)
			if err != nil {
				log.Printf("WARN Couldn't decode URL %s, %s", b, err)
			}
			hl[h] = append(hl[h], *e)
		}
	}

	return &hl, nil
}

func (hb *HipacheBackend) getFrontends() ([]string, error) {
	r, err := redis.Values(hb.r.Do("KEYS", "frontend:*"))
	if err != nil {
		return nil, err
	}

	var fe []string
	if err := redis.ScanSlice(r, &fe); err != nil {
		return nil, err
	}

	return fe, nil
}

func (hb *HipacheBackend) hostExists(h shared.Host) (bool, error) {
	return redis.Bool(hb.r.Do("EXISTS", prefixKey(h)))
}

func (hb *HipacheBackend) hostDelete(h shared.Host) error {
	if _, err := hb.r.Do("DEL", prefixKey(h)); err != nil {
		return err
	}
	log.Printf("DEBUG [backend:hipache] Host deleted '%s'\n", h)

	return nil
}

func (hb *HipacheBackend) hostCreate(h shared.Host) error {
	if _, err := hb.r.Do("RPUSH", prefixKey(h), h); err != nil {
		return err
	}
	log.Printf("DEBUG [backend:hipache] Host created: %s\n", h)

	return nil
}

func (hb *HipacheBackend) clearHosts() error {
	fe, err := hb.getFrontends()
	if err != nil {
		return err
	}

	for _, f := range fe {
		if _, err := hb.r.Do("DEL", f); err != nil {
			return err
		}
	}

	return nil
}

func prefixKey(h shared.Host) string {
	return "frontend:" + string(h)
}

func init() {
	backends.BackendMap["hipache"] = NewHipacheBackend
}
