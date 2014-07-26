package hipdate

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
type Host string

func (h Host) Key() string {
	return "frontend:" + string(h)
}

type UpstreamList map[Upstream]int
type HostList map[Host]UpstreamList

func (hl HostList) Add(h Host, u Upstream) {
	if _, exists := hl[h]; !exists {
		hl[h] = UpstreamList{}
	}
	hl[h][u] = 0
}

func (hl HostList) Remove(h Host, u Upstream) {
	delete(hl[h], u)
	if len(hl[h]) == 0 {
		delete(hl, h)
	}
}
