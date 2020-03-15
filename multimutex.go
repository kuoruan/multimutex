package multimutex

import (
	"errors"
	"sync"
)

/** refs:
 *	- https://gist.github.com/fzerorubigd/fe429d063d95cbe44ec3
 *  - https://github.com/im7mortal/kmutex
 */

var (
	defaultMutexPool = &sync.Pool{
		New: func() interface{} {
			return new(sync.RWMutex)
		},
	}
	defaultLockerMap = make(map[interface{}]*mLocker)
)

type MultiLocker interface {
	Lock(interface{})
	Unlock(interface{})
	RLock(interface{})
	RUnlock(interface{})
}

type mLocker struct {
	count int64
	mux   *sync.RWMutex
}

type lockerMap map[interface{}]*mLocker

type MultiMutex struct {
	mux     sync.Mutex
	pool    *sync.Pool
	lockers lockerMap
}

func (m *MultiMutex) Lock(id interface{}) {
	ml := m.getOrNewLocker(id)
	ml.count++

	ml.mux.Lock()
}

func (m *MultiMutex) Unlock(id interface{}) {
	ml := m.getLocker(id)

	if ml != nil {
		ml.mux.Unlock()
		m.releaseLocker(id, ml)
	}
}

func (m *MultiMutex) RLock(id interface{}) {
	ml := m.getOrNewLocker(id)
	ml.count++

	ml.mux.RLock()
}

func (m *MultiMutex) RUnlock(id interface{}) {
	ml := m.getLocker(id)

	if ml != nil {
		ml.mux.RUnlock()
		m.releaseLocker(id, ml)
	}
}

func (m *MultiMutex) getLocker(id interface{}) *mLocker {
	m.mux.Lock()
	defer m.mux.Unlock()

	lockerMap := m.getLockerMap()
	return lockerMap[id]
}

func (m *MultiMutex) getOrNewLocker(id interface{}) *mLocker {
	m.mux.Lock()
	defer m.mux.Unlock()

	lockerMap := m.getLockerMap()
	ml, has := lockerMap[id]

	if !has {
		mutexPool := m.getMutexPool()
		lockerMap := m.getLockerMap()

		l := mutexPool.Get()
		lock, ok := l.(*sync.RWMutex)
		if !ok {
			panic(errors.New("the pool returns invalid value"))
		}
		ml = &mLocker{
			count: 0,
			mux:   lock,
		}

		lockerMap[id] = ml
	}

	return ml
}

func (m *MultiMutex) releaseLocker(id interface{}, locker *mLocker) {
	m.mux.Lock()
	defer m.mux.Unlock()

	locker.count--
	if locker.count <= 0 {
		mutexPool := m.getMutexPool()
		lockerMap := m.getLockerMap()

		mutexPool.Put(locker.mux)

		delete(lockerMap, id)
	}
}

func (m *MultiMutex) getMutexPool() *sync.Pool {
	if m.pool != nil {
		return m.pool
	}

	return defaultMutexPool
}

func (m *MultiMutex) getLockerMap() lockerMap {
	if m.lockers != nil {
		return m.lockers
	}
	return defaultLockerMap
}

func New() *MultiMutex {
	return &MultiMutex{
		mux: sync.Mutex{},
		pool: &sync.Pool{
			New: func() interface{} {
				return new(sync.RWMutex)
			},
		},
		lockers: make(map[interface{}]*mLocker),
	}
}
