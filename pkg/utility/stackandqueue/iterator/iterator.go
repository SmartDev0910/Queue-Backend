package iterator

import (
	"github.com/CS-PCockrill/queue/pkg/utility/stackandqueue/node"
)

type IIterator interface {
	HasNext() bool
	Next() *node.Node
}
