package utils

import "sync"

var CopyBufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 4096)
	},
}
