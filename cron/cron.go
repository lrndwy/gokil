package cron

import (
	"context"
	"log"
	"sync"
	"time"
)

type Job struct {
	Name       string
	Every      time.Duration
	Run        func(context.Context) error
	RunOnStart bool
}

type Runner struct {
	Jobs    []Job
	OnError func(job Job, err error)
	Logger  *log.Logger
}

// Run blocks until ctx is cancelled.
func (r Runner) Run(ctx context.Context) error {
	if r.OnError == nil {
		r.OnError = func(job Job, err error) {
			if r.Logger != nil {
				r.Logger.Printf("cron job %q error: %v", job.Name, err)
			}
		}
	}

	var wg sync.WaitGroup
	for _, job := range r.Jobs {
		j := job
		if j.Every <= 0 || j.Run == nil {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()

			if j.RunOnStart {
				if err := j.Run(ctx); err != nil {
					r.OnError(j, err)
				}
			}

			ticker := time.NewTicker(j.Every)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if err := j.Run(ctx); err != nil {
						r.OnError(j, err)
					}
				}
			}
		}()
	}

	<-ctx.Done()
	wg.Wait()
	return nil
}

// Run is the simplest API: cron.Run(ctx, jobs...)
func Run(ctx context.Context, jobs ...Job) error {
	return Runner{Jobs: jobs}.Run(ctx)
}

