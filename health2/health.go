package health2

type Option func(*health)

func DefineStat(key string, s Stat) Option {
	return func(h *health) {
		h.stats[key] = s
	}
}

type Interface interface {
	Stat(key string) Stat
}

type health struct {
	stats map[string]Stat
}

func (h *health) Stat(key string) Stat {
	return h.stats[key]
}

func New(o ...Option) Interface {
	h := &health{
		stats: make(map[string]Stat),
	}

	for _, option := range o {
		option(h)
	}

	return h
}
