package queue

import (
	"container/list"
	"sync"
)

type SyncQueue struct {
	sync.RWMutex
	*list.List
}

func NewSyncQueue() *SyncQueue {
	return &SyncQueue{
		List: list.New(),
	}
}

func (s *SyncQueue) Len() int {
	s.RLock()
	defer s.RUnlock()
	return s.List.Len()
}

func (s *SyncQueue) Enqueue(i interface{}) {
	s.Lock()
	defer s.Unlock()
	s.List.PushBack(i)
}

func (s *SyncQueue) Dequeue() interface{} {
	s.Lock()
	defer s.Unlock()
	e := s.List.Front()
	if e != nil {
		s.List.Remove(e)
		return e.Value
	}
	return nil
}