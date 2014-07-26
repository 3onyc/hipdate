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
