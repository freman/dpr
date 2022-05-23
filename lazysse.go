package main

import (
	"sync"

	"github.com/freman/sse"
)

type lazySSE struct {
	*sse.EventStream
	es func() *sse.EventStream
	mu sync.Mutex
}

func newLazySSE() *lazySSE {
	tmp := &lazySSE{}
	tmp.mu.Lock()

	tmp.es = func() *sse.EventStream {
		defer tmp.mu.Unlock()
		if tmp.EventStream == nil {
			tmp.EventStream = sse.New()
		}

		tmp.es = func() *sse.EventStream {
			return tmp.EventStream
		}

		return tmp.EventStream
	}

	return tmp
}
