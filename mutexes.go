package mutexes

import (
	"runtime"
	"sync"
	"sync/atomic"
)

type Mutexes struct {
	lockers sync.Map
	mutex   sync.RWMutex
}

func (m *Mutexes) Get(key interface{}) sync.Locker {
	m.mutex.RLock()
	l, _ := m.lockers.LoadOrStore(key, &locker{key: key})
	lw := newLockerWrapper(m, l.(*locker))
	m.mutex.RUnlock()
	runtime.SetFinalizer(lw, lockerFinalizer)
	return lw
}

func (m *Mutexes) Lock(key interface{}) {
	m.Get(key).Lock()
}

func (m *Mutexes) Unlock(key interface{}) {
	m.Get(key).Unlock()
}

type locker struct {
	key   interface{}
	mutex sync.Mutex
	ref   int32
}

var _ sync.Locker = (*locker)(nil)

func (l *locker) Lock() {
	atomic.AddInt32(&l.ref, 1)
	l.mutex.Lock()
}

func (l *locker) Unlock() {
	l.mutex.Unlock()
	atomic.AddInt32(&l.ref, -1)
}

type lockerWrapper struct {
	*locker
	mutexes *Mutexes
}

func newLockerWrapper(mutexes *Mutexes, l *locker) *lockerWrapper {
	atomic.AddInt32(&l.ref, 1)
	return &lockerWrapper{locker: l, mutexes: mutexes}
}

func lockerFinalizer(lw *lockerWrapper) {
	ref := atomic.AddInt32(&lw.ref, -1)
	if ref == 0 {
		m := lw.mutexes
		m.mutex.Lock()
		if atomic.LoadInt32(&lw.ref) == 0 {
			m.lockers.Delete(lw.key)
		}
		m.mutex.Unlock()
	} else if ref < 0 {
		panic("mutexes: Reference count is negative")
	}
}
