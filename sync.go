package phly

import (
	"container/list"
	"github.com/micro-go/lock"
	"sync"
)

// --------------------------------
// SHUTTLE

// shuttle is a thread-safe list of items.
type shuttle struct {
	mutex sync.Mutex
	list  list.List
}

func (s *shuttle) len() int {
	defer lock.Locker(&s.mutex).Unlock()
	return s.list.Len()
}

func (s *shuttle) push(i interface{}) {
	if i == nil {
		return
	}
	defer lock.Locker(&s.mutex).Unlock()
	s.list.PushBack(i)
}

func (s *shuttle) pop() (interface{}, error) {
	defer lock.Locker(&s.mutex).Unlock()
	e := s.list.Front()
	if e == nil {
		return nil, emptyErr
	}
	s.list.Remove(e)
	return e.Value, nil
}
