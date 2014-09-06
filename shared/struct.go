package shared

import (
	"bytes"
	"hash/crc32"
	"strconv"
)

type Stoppable struct {
	stopChan chan bool
}

func (s Stoppable) ShouldStop() bool {
	select {
	case <-s.stopChan:
		return true
	default:
		return false
	}
}

func (s Stoppable) setChan(c chan bool) {
	s.stopChan = c
}

type OptionMap map[string]string
type HostList map[Host][]Upstream

func (hl HostList) Pprint() string {
	buf := bytes.Buffer{}

	buf.WriteString("<h1>Hosts</h1>" + "\n<ul>")
	for h, bs := range hl {
		buf.WriteString("<li>" + string(h) + "\n<ul>\n")
		for _, b := range bs {
			buf.WriteString("<li>" + string(b) + "</li>\n")
		}
		buf.WriteString("</ul>\n</li>\n")
	}
	buf.WriteString("</ul>\n")

	return buf.String()
}

type ChangeEvent struct {
	Type string
	Host Host
	IP   IPAddress
}

func NewChangeEvent(t string, h Host, ip IPAddress) *ChangeEvent {
	return &ChangeEvent{
		Type: t,
		Host: h,
		IP:   ip,
	}
}

type IPAddress string
type ContainerID string
type Upstream string

func (u Upstream) Hash() string {
	return strconv.FormatUint(uint64(crc32.ChecksumIEEE([]byte(u))), 10)
}

type Host string
