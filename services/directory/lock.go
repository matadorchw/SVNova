// Copyright 2026 The SVNova Authors. All rights reserved.
// SPDX-License-Identifier: MIT
package directory

import (
	"context"
	"sync"
)

// A channel queues waiting callers and allows scheduled work to cancel during shutdown.
type directoryMutex struct {
	once sync.Once
	gate chan struct{}
}

func (m *directoryMutex) channel() chan struct{} {
	m.once.Do(func() { m.gate = make(chan struct{}, 1) })
	return m.gate
}
func (m *directoryMutex) Lock() { m.channel() <- struct{}{} }
func (m *directoryMutex) Unlock() {
	select {
	case <-m.channel():
	default:
		panic("unlock of unlocked directory mutex")
	}
}
func (m *directoryMutex) TryLock() bool {
	select {
	case m.channel() <- struct{}{}:
		return true
	default:
		return false
	}
}
func (m *directoryMutex) LockContext(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	select {
	case m.channel() <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
