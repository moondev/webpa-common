package health2

import (
	"sync"
)

type Interface interface {
}

type health struct {
	mutex sync.RWMutex
}

func New() Interface {
	return nil
}
