package utils

import "sync"

type SafeMap struct {
	sync.RWMutex
	Map map[interface{}]interface{}
}

func NewSafeMap(size int) *SafeMap {
	sm := new(SafeMap)
	sm.Map = make(map[interface{}]interface{}, size)
	return sm
}

//为提高性能，读不加锁
func (sm *SafeMap) ReadMap(key interface{}) interface{} {
	value := sm.Map[key]
	return value
}

func (sm *SafeMap) WriteMap(key interface{}, value interface{}) {
	sm.Lock()
	sm.Map[key] = value
	sm.Unlock()
}
