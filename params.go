package hxlru

import "time"

type ParamsNewCacheLRU struct {
	TTL      time.Duration
	Capacity uint16
}
