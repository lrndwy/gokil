package cron_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lrndwy/gokil/cron"
)

func TestRunnerRunsJob(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var ran int64
	job := cron.Job{
		Name:       "ping",
		Every:      10 * time.Millisecond,
		RunOnStart: true,
		Run: func(context.Context) error {
			atomic.AddInt64(&ran, 1)
			return nil
		},
	}

	done := make(chan struct{})
	go func() {
		_ = cron.Run(ctx, job)
		close(done)
	}()

	time.Sleep(35 * time.Millisecond)
	cancel()

	<-done
	if atomic.LoadInt64(&ran) == 0 {
		t.Fatal("expected job to run at least once")
	}
}

