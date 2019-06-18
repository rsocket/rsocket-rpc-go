package common

import (
	"bytes"
	"sync"
)

var bPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func BorrowBuffer() (b *bytes.Buffer) {
	b = bPool.Get().(*bytes.Buffer)
	b.Reset()
	return
}

func ReturnBuffer(b *bytes.Buffer) {
	bPool.Put(b)
}
