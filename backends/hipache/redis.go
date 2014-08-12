package hipache

import (
	"errors"
	"github.com/garyburd/redigo/redis"
	"net/url"
)

var (
	WrongSchemeError = errors.New("scheme is not redis://")
)

func parseRedisUrl(urlStr string) (string, error) {
	redisUrl, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	if redisUrl.Scheme != "redis" {
		return "", WrongSchemeError
	}

	return redisUrl.Host, nil
}

func createRedisConn(ru string) (*redis.Conn, error) {
	re, err := parseRedisUrl(ru)
	if err != nil {
		return nil, err
	}

	r, err := redis.Dial("tcp", re)
	if err != nil {
		return nil, err
	}

	return &r, nil
}
