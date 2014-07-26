package hipdate

import (
	"errors"
	"net/url"
)

func ParseRedisUrl(urlStr string) (string, error) {
	redisUrl, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	if redisUrl.Scheme != "redis" {
		return "", errors.New("Scheme is not redis://")
	}

	return redisUrl.Host, nil
}
