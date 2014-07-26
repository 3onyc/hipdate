package hipdate

type Backend interface {
	AddUpstream(h Host, u Upstream) error
	RemoveUpstream(h Host, u Upstream) error
}

type Source interface {
	Start()
	Stop()
}
