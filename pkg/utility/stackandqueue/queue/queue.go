package queue

import (
	"github.com/CS-PCockrill/queue/pkg/utility/stackandqueue/linkedList"
	"github.com/CS-PCockrill/queue/pkg/utility/stackandqueue/node"
	it "github.com/CS-PCockrill/queue/pkg/utility/stackandqueue/iterator"
)

type IQueue interface {
	Enqueue(item int) *node.Node
	Dequeue() *node.Node
	Peek() *node.Node
	Iterator() it.IIterator
}

type Queue struct {
	li *linkedList.LinkedList
}

func New() *Queue {
	list := linkedList.New(nil, nil)
	return &Queue{list}
}

func (q *Queue) Iterator() it.IIterator {
	return q.li.Iterator()
}

func (q *Queue) Peek() *node.Node {
	return q.li.Head()
}

func (q *Queue) Enqueue(item int) *node.Node {
	newNode := q.li.Append(item)
	return newNode
}

func (q *Queue) Dequeue() *node.Node {
	removedNode := q.li.RemoveFront()
	return removedNode
}

