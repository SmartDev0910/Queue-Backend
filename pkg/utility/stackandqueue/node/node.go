package node

type Node struct {
	item int
	next *Node
}

func New(item int) *Node {
	return &Node{item, nil}
}

func (n *Node) Item() int {
	return n.item
}

func (n *Node) Next() *Node {
	return n.next
}

func (n *Node) SetItem(item int) int {
	n.item = item
	return item
}

func (n *Node) SetNext(nextNode *Node) *Node {
	n.next = nextNode
	return n.next
}
