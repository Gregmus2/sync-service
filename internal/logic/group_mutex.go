package logic

import "sync"

type groupMutex struct {
	mutexes map[string]*sync.Mutex
}

func NewGroupMutex() GroupMutex {
	return &groupMutex{
		mutexes: make(map[string]*sync.Mutex),
	}
}

func (g groupMutex) Lock(groupID string) {
	if _, ok := g.mutexes[groupID]; !ok {
		g.mutexes[groupID] = &sync.Mutex{}
	}

	g.mutexes[groupID].Lock()
}

func (g groupMutex) Unlock(groupID string) {
	g.mutexes[groupID].Unlock()
}
