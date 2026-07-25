package models

import (
	"context"
	"runtime"
	"sync"
)

var contextStore sync.Map

func SetContext(ctx context.Context) {
	contextStore.Store(gid(), ctx)
}

func GetContext() context.Context {
	if ctx, ok := contextStore.Load(gid()); ok {
		return ctx.(context.Context)
	}
	return context.Background()
}

func ClearContext() {
	contextStore.Delete(gid())
}

func gid() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	// stack starts with "goroutine <id> "
	const prefix = "goroutine "
	if n < len(prefix)+1 {
		return 0
	}
	var id uint64
	for i := len(prefix); i < n; i++ {
		c := buf[i]
		if c == ' ' {
			break
		}
		if c < '0' || c > '9' {
			break
		}
		id = id*10 + uint64(c-'0')
	}
	return id
}
