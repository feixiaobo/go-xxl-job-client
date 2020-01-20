package utils

import (
	"errors"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
)

type Node struct {
	Item interface{}
	Next *Node
}

type Queue struct {
	sync.Mutex
	Count    int32
	Capacity int32
	Head     *Node
	Last     *Node
}

func NewQueue() *Queue {
	node := &Node{}
	return &Queue{
		Count:    int32(0),
		Capacity: math.MaxInt32,
		Head:     node,
		Last:     node,
	}
}

func (q *Queue) Put(item interface{}) error {
	if item == nil {
		return errors.New("item can't be nil")
	}

	getPos := atomic.LoadInt32(&q.Count)
	if getPos >= q.Capacity {
		return errors.New("queue size exceeding maximum capacity")
	}

	node := &Node{Item: item}
	q.Lock()
	defer q.Unlock()
	atomic.AddInt32(&q.Count, 1)
	q.Last.Next = node
	q.Last = node
	return nil
}

//保证单协程执行，不加锁
func (q *Queue) Poll() (has bool, item interface{}) {
	getPos := atomic.LoadInt32(&q.Count)
	if getPos <= int32(0) {
		runtime.Gosched()
		return false, nil
	}
	atomic.AddInt32(&q.Count, -1)
	node := q.Head.Next
	res := node.Item
	node.Item = nil //help GC
	q.Head = node
	return true, res
}

func (q *Queue) Clear() {
	q = NewQueue()
}
