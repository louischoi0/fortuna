package service

import (
	"sync"
)

type LoadBalancer struct {
	mu	sync.Mutex
}

