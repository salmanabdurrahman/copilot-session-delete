package deletion

import (
	"context"
	"os"
)

// Executor carries out deletion plans produced by Planner.
type Executor struct {
	dryRun bool
}

// NewExecutor creates an Executor. When dryRun is true no files are touched.
func NewExecutor(dryRun bool) *Executor {
	return &Executor{dryRun: dryRun}
}

// Execute runs the plans sequentially and streams each Result to the returned
// channel. The channel is closed when all plans have been processed or the
// context is cancelled.
//
// Deletion is sequential (not concurrent) to avoid partial-delete races on
// disk. Cancellation is honoured between sessions, not in the middle of a
// single os.RemoveAll call.
func (e *Executor) Execute(ctx context.Context, plans []Plan) <-chan Result {
	ch := make(chan Result, len(plans))
	go func() {
		defer close(ch)
		for _, p := range plans {
			if ctx.Err() != nil {
				return
			}
			ch <- e.executeOne(p)
		}
	}()
	return ch
}

func (e *Executor) executeOne(p Plan) Result {
	r := Result{
		SessionID: p.Session.ID,
		DryRun:    e.dryRun,
	}

	if p.NotFound {
		r.Success = true
		r.NotFound = true
		return r
	}

	if e.dryRun {
		r.Success = true
		return r
	}

	if err := os.RemoveAll(p.Session.RootPath); err != nil {
		r.Success = false
		r.Err = err
		return r
	}

	r.Success = true
	return r
}
