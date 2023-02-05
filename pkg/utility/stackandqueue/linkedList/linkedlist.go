package linkedList

import (
	"fmt"
	it "github.com/CS-PCockrill/queue/pkg/utility/stackandqueue/iterator"
	"github.com/CS-PCockrill/queue/pkg/utility/stackandqueue/node"
)

type iterator struct {
	current *node.Node
}

func (it *iterator) Next() *node.Node {
	current := it.current
	it.current = it.current.Next()
	return current
}

func (it *iterator) HasNext() bool {
	if it.current == nil {
		return false
	}
	return true
}

type LinkedList struct {
	head *node.Node
	current *node.Node
	length int
}

func (li *LinkedList) Iterator() it.IIterator {
	head := li.head
	return &iterator{head}
}

func New(head, current *node.Node) *LinkedList {
	return &LinkedList{head, current, 0}
}

func (li *LinkedList) Head() *node.Node {
	return li.head
}

func (li *LinkedList) Print() {
	head := li.head
	for head != nil {
		fmt.Println("%v\n", head.Item())
		head = head.Next()
	}
}

func (li *LinkedList) Insert(item int) *node.Node {
	newNode := node.New(item)
	if li.head == nil {
		li.head = newNode
		li.current = newNode
		li.length++
		return newNode
	}
	newNode.SetNext(li.current)
	li.current = newNode
	li.head = newNode
	li.length = li.length + 1
	return newNode
}

func (li *LinkedList) Append(item int) *node.Node {
	newNode := node.New(item)
	if li.head == nil {
		li.head = newNode
		li.current = newNode
		li.length++
		return newNode
	}

	li.current.SetNext(newNode)
	li.current = newNode
	li.length++
	return newNode
}

func (li *LinkedList) RemoveFront() *node.Node {
	head := li.head
	if head != nil {
		nextNode := head.Next()
		li.head = nextNode
		head = nil
	}
	return nil
}
