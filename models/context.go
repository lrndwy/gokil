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
	var id uint64
	for s := buf[6]; s < byte(n); s++ {
		if buf[s] == ' ' {
			break
		}
		id = id*10 + uint64(buf[s]-'0')
	}
	return id
}
