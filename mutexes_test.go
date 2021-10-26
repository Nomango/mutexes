package mutexes_test

import (
	"runtime"
	"testing"
	"time"

	"github.com/Nomango/mutexes"
)

func TestMutexesDrop(t *testing.T) {
	var m mutexes.Mutexes

	m.Get("test").Lock()
	time.Sleep(time.Millisecond * 20)
	runtime.GC()
	time.Sleep(time.Millisecond * 20)

	m.Get("test").Unlock()
	time.Sleep(time.Millisecond * 20)
	runtime.GC()
	time.Sleep(time.Millisecond * 20)
}

func BenchmarkMutexes(b *testing.B) {
	var (
		m mutexes.Mutexes
		i int64
	)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		l := m.Get("test")
		for pb.Next() {
			l.Lock()
			i++
			l.Unlock()
		}
	})
}

func BenchmarkMutexes2(b *testing.B) {
	var (
		m mutexes.Mutexes
		i int64
	)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Lock("test")
			i++
			m.Unlock("test")
		}
	})
}
