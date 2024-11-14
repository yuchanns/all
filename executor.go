/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package all provides a parallelism tool library for executing tasks with a specified number of workers.
package all

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/panjf2000/ants/v2"
	"github.com/smallnest/chanx"
)

type ordered[T any] struct {
	index uint
	data  T
	ctx   context.Context
}

type executor[T any, R any] struct {
	input  *chanx.UnboundedChan[ordered[T]]
	output *chanx.UnboundedChan[ordered[R]]
	errors *chanx.UnboundedChan[error]

	abortOnError bool

	ctx         context.Context
	isClosed    int32
	assignIndex uint64

	result R
	index  uint
	err    error
}

// NewVoid creates a new executor that processes tasks without returning any result.
//
// Parameters:
// - ctx: The context for managing the lifecycle of the executor.
// - consumer: The function to process each task.
// - workers: Optional number of workers to process tasks concurrently.
//
// Returns:
// - A pointer to the created executor.
func NewVoid[T any](ctx context.Context, consumer func(context.Context, T) error, workers ...uint) *executor[T, struct{}] {
	return New(ctx, func(ctx context.Context, item T) (struct{}, error) {
		return struct{}{}, consumer(ctx, item)
	}, workers...)
}

// New creates a new executor that processes tasks and returns results.
//
// Parameters:
// - ctx: The context for managing the lifecycle of the executor.
// - consumer: The function to process each task and return a result.
// - workers: Optional number of workers to process tasks concurrently.
//
// Returns:
// - A pointer to the created executor.
func New[T any, R any](ctx context.Context, consumer func(context.Context, T) (R, error), workers ...uint) *executor[T, R] {
	var workerNum int = 16
	if len(workers) > 0 {
		workerNum = int(workers[0])
	}

	x := &executor[T, R]{
		input:        chanx.NewUnboundedChan[ordered[T]](ctx, workerNum),
		output:       chanx.NewUnboundedChan[ordered[R]](ctx, workerNum),
		errors:       chanx.NewUnboundedChan[error](ctx, 1),
		ctx:          ctx,
		abortOnError: true,
	}

	pool, _ := ants.NewPool(workerNum)

	go func() {
		defer func() {
			pool.Release()
			close(x.errors.In)
			close(x.output.In)
		}()

		wg := sync.WaitGroup{}

		for item := range x.input.Out {
			wg.Add(1)
			pool.Submit(func() {
				var err error

				defer func() {
					if msg := recover(); msg != nil {
						err = fmt.Errorf("panic: %s", msg)
					}
					if err != nil {
						x.errors.In <- err
						if x.abortOnError {
							pool.Release()
						}
					}
					wg.Done()
				}()

				data, err := consumer(item.ctx, item.data)
				if err != nil {
					return
				}
				v := reflect.ValueOf(data)
				if (v.Kind() == reflect.Ptr || v.Kind() == reflect.Map) && v.IsNil() {
					return
				}
				x.output.In <- ordered[R]{index: item.index, data: data}
			})
		}
		wg.Wait()
	}()

	return x
}

func (x *executor[T, R]) assign(ctx context.Context, item T, index uint) {
	if atomic.LoadInt32(&x.isClosed) == 1 {
		return
	}
	x.input.In <- ordered[T]{index: index, data: item, ctx: ctx}
}

// Assign adds a task to the executor.
//
// Parameters:
// - item: The task to be processed.
func (x *executor[T, R]) Assign(item T) {
	index := atomic.AddUint64(&x.assignIndex, 1)
	x.assign(x.ctx, item, uint(index))
}

// AssignWithContext adds a task to the executor with a specific context.
//
// Parameters:
// - ctx: The context for the task.
// - item: The task to be processed.
func (x *executor[T, R]) AssignWithContext(ctx context.Context, item T) {
	index := atomic.AddUint64(&x.assignIndex, 1)
	x.assign(ctx, item, uint(index))
}

// Next retrieves the next result from the executor.
//
// You won't be able to assign new tasks after calling Next.
//
// Returns:
// - ok: A boolean indicating whether a result was retrieved.
func (x *executor[T, R]) Next() (ok bool) {
	if atomic.CompareAndSwapInt32(&x.isClosed, 0, 1) {
		close(x.input.In)
	}
	result, ok := <-x.output.Out
	if ok {
		x.result = result.data
		x.index = result.index
	}
	return
}

// Each returns the current result from the executor.
//
// Returns:
// - result: The current result.
func (x *executor[T, R]) Each() (result R) {
	if x.err != nil && x.abortOnError {
		return
	}
	result = x.result
	return
}

// Index returns the index of the current result in relation to the order of assigned tasks.
//
// Returns:
// - The index of the current result.
func (x *executor[T, R]) Index() uint {
	if x.err != nil && x.abortOnError {
		return 0
	}
	return x.index
}

// Error returns any error encountered during the execution of tasks.
//
// Returns:
// - The error encountered, or nil if no error occurred.
func (x *executor[T, R]) Error() error {
	if x.err == nil {
		x.err = context.Cause(x.ctx)
	}
	if x.err == nil {
		errs := &errors{}
		for e := range x.errors.Out {
			errs.errs = append(errs.errs, e)
		}
		if len(errs.errs) > 0 {
			x.err = errs
		}
	}
	return x.err
}
