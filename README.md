# mutexes

Usage:

```golang
var m mutexes.Mutexes

locker := m.Get("test")
locker.Lock()
defer locker.Unlock()
```

```golang
// 不推荐
var m mutexes.Mutexes
m.Lock("test")
defer m.Unlock("test")
```
