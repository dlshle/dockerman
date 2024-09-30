package async

import (
	"context"
	"sync/atomic"
)

// AsyncGroup[T] maintains a group of asynchronous tasks. It ensures that when
// any of the tasks completes, the group stops accepting new tasks and stores
// the result for retrieval.
type AsyncGroup[T any] struct {
	ctx       context.Context
	ctxCloser func()
	closed    *atomic.Bool
	result    *atomic.Value
}

func NewAsyncGroup[T any]() *AsyncGroup[T] {
	ctx, cancelFunc := context.WithCancel(context.Background())
	return &AsyncGroup[T]{
		ctx:       ctx,
		ctxCloser: cancelFunc,
		closed:    &atomic.Bool{},
		result:    &atomic.Value{},
	}
}

func (g *AsyncGroup[T]) Run(f func(context.Context) T) {
	if g.closed.Load() {
		return
	}
	go func() {
		result := make(chan T, 1)
		select {
		case result <- f(g.ctx):
			g.closed.Store(true)
			g.result.Store(<-result)
			g.ctxCloser()
		case <-g.ctx.Done():
			close(result)
		}
	}()
}

func (g *AsyncGroup[T]) Race() T {
	<-g.ctx.Done()
	return g.result.Load().(T)
}
