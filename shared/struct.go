package shared

import (
	"bytes"
	"errors"
	"fmt"
	"hash/crc32"
	"net/url"
	"strconv"
	"strings"
)

type OptionMap map[string]string
type HostList map[Host][]Endpoint

func (hl HostList) Pprint() string {
	buf := bytes.Buffer{}

	buf.WriteString("<h1>Hosts</h1>" + "\n<ul>")
	for h, eps := range hl {
		buf.WriteString("<li>" + string(h) + "\n<ul>\n")
		for _, ep := range eps {
			buf.WriteString("<li>" + ep.String() + "</li>\n")
		}
		buf.WriteString("</ul>\n</li>\n")
	}
	buf.WriteString("</ul>\n")

	return buf.String()
}

type ChangeEvent struct {
	Type     string
	Host     Host
	Endpoint Endpoint
}

func NewChangeEvent(t string, h Host, e Endpoint) *ChangeEvent {
	return &ChangeEvent{
		Type:     t,
		Host:     h,
		Endpoint: e,
	}
}

type Endpoint struct {
	Scheme  string
	Address string
	Port    uint32
}

func (e *Endpoint) String() string {
	return fmt.Sprintf("%s://%s:%d", e.Scheme, e.Address, e.Port)
}

func (e *Endpoint) Hash() string {
	return strconv.FormatUint(uint64(crc32.ChecksumIEEE([]byte(e.String()))), 10)
}

func NewEndpoint(s, a string, p uint32) *Endpoint {
	return &Endpoint{
		Scheme:  s,
		Address: a,
		Port:    p,
	}
}

func NewEndpointFromUrl(us string) (*Endpoint, error) {
	u, err := url.Parse(us)
	if err != nil {
		return nil, err
	}

	if !strings.Contains(u.Host, ":") {
		return nil, errors.New("Missing port in URL")
	}

	s := strings.SplitN(u.Host, ":", 2)
	p, err := strconv.ParseUint(s[1], 10, 32)
	if err != nil {
		return nil, err
	}

	return NewEndpoint(u.Scheme, s[0], uint32(p)), nil
}

type ContainerID string

type Host string
