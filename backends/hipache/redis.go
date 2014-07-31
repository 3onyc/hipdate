package hipache

import (
	"errors"
	"net/url"
)

var (
	WrongSchemeError = errors.New("Scheme is not redis://")
)

func ParseRedisUrl(urlStr string) (string, error) {
	redisUrl, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	if redisUrl.Scheme != "redis" {
		return "", WrongSchemeError
	}

	return redisUrl.Host, nil
}
