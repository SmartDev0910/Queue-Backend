package queue

import (
	"testing"
	"fmt"
)

func TestQueue(t *testing.T) {
	var queue IQueue = New()
	queue.Enqueue(32)
	queue.Enqueue(40)
	queue.Enqueue(100)

	peek := queue.Peek().Item()
	if peek != 32 {
		t.Errorf("got %v | expected 32 \n", peek)
	}
	iter := queue.Iterator()
	for iter.HasNext() {
		fmt.Printf("item %v \n", iter.Next().Item())
	}
}

func TestDequeue(t *testing.T) {
	var queue IQueue = New()
	queue.Enqueue(32)
	queue.Enqueue(40)
	queue.Enqueue(100)
	queue.Enqueue(420)
	queue.Enqueue(312)
	queue.Enqueue(2)

	queue.Dequeue()

	peek := queue.Peek()
	if peek.Item() != 40 {
		fmt.Printf("item %v \n", peek.Item())
		t.Errorf("wrong dequeue")
	}
	queue.Dequeue()
	queue.Dequeue()

	fmt.Println("====== Start iterator Dequeue ======")
	iter := queue.Iterator()
	for iter.HasNext() {
		fmt.Printf("item %v \n", iter.Next().Item())
	}
	fmt.Println("======= End iterator Dequeue =======")
	peek = queue.Peek()
	if peek.Item() != 2 {
		t.Errorf("got %v | expected 2 \n", peek.Item())
		t.Errorf("wrong dequeue")
	}
}

