package mutexes

import (
	"runtime"
	"sync"
	"sync/atomic"
)

type Mutexes struct {
	mutex   sync.Mutex
	lockers map[interface{}]*locker
}

func (m *Mutexes) Get(key interface{}) sync.Locker {
	m.mutex.Lock()
	if m.lockers == nil {
		m.lockers = make(map[interface{}]*locker)
	}
	l := m.lockers[key]
	if l == nil {
		l = &locker{key: key}
		m.lockers[key] = l
	}
	// 创建 LockerWrapper 保证资源及时回收
	lw := newLockerWrapper(m, l)
	runtime.SetFinalizer(lw, lockerFinalizer)
	m.mutex.Unlock()
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
	count int32
}

var _ sync.Locker = (*locker)(nil)

func (l *locker) Lock() {
	atomic.AddInt32(&l.count, 1)
	l.mutex.Lock()
}

func (l *locker) Unlock() {
	l.mutex.Unlock()
	atomic.AddInt32(&l.count, -1)
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
	if ref < 0 {
		panic("mutexes: Reference count is negative")
	}
	if ref == 0 && atomic.LoadInt32(&lw.count) == 0 {
		m := lw.mutexes
		m.mutex.Lock()
		if atomic.LoadInt32(&lw.ref) == 0 {
			delete(m.lockers, lw.key)
		}
		m.mutex.Unlock()
	}
}
